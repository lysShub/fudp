package ecc

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"errors"
	"math/big"
)

/* 实现ECC_256的加密解密、签名验签; 用于建立安全的信道  */
// 	ECC原始功能不包含加密解密功能、只包含签名验签功能; 加密是借助AES_GCM_256完成的。
//
// 	公钥的序列化采用elliptic(ANSI X9.62:4.3.6)
// 	私钥的序列化采用x509, 因为不参与通信、其他语言怎么实现都行

//	与RSA不同,
// 	因此、对于基于RSA的TLS的Client回复密钥时只包含公钥加密后的密钥; 但基于ECC时, 不仅包含加密后的密钥
// 	还包括Client的公钥, 加密后密钥与Client公钥拼接成实公钥。
//
// 	实公钥的组成结构为

func GenerateKey() (priKey, pubKey []byte, err error) {
	var key *ecdsa.PrivateKey
	if key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err != nil {
		priKey, pubKey = nil, nil
	} else {

		if priKey, err = x509.MarshalECPrivateKey(key); err != nil {
			priKey, pubKey = nil, nil
			return
		}
		pubKey = elliptic.MarshalCompressed(elliptic.P256(), key.X, key.Y)
	}
	return
}

// Encrypt 公钥加密
// 	密文结构为：{Client公钥 加密后密文 Client公钥长度(1B,单位字节)}
func Encrypt(pubkey []byte, plaintext []byte) (ciphertext []byte, err error) {

	var selfKey *ecdsa.PrivateKey
	if selfKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err != nil {
		return nil, err
	}
	selfPub := elliptic.MarshalCompressed(elliptic.P256(), selfKey.X, selfKey.Y)

	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), pubkey)
	if x == nil || y == nil {
		return nil, errors.New("invalid ECC public key")
	} else {
		if ok := elliptic.P256().IsOnCurve(x, y); !ok {
			return nil, errors.New("parse ECC publice key with ELLIPTIC(ANSI X9.62) error")
		}
	}
	var sps ecdsa.PublicKey = ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
	sx, _ := sps.Curve.ScalarMult(x, y, selfKey.D.Bytes())
	if sx == nil {
		return nil, errors.New("encrypt fail")
	}
	shard := sha256.Sum256(sx.Bytes()) // AES_GCM_256对称加密密钥

	// 加密
	ciphertext = make([]byte, 0, len(plaintext)+16+33+1)
	var gcm cipher.AEAD
	if block, err := aes.NewCipher(shard[:]); err != nil {
		return nil, err
	} else {
		if gcm, err = cipher.NewGCM(block); err != nil {
			return nil, err
		}
	}
	ciphertext = gcm.Seal(ciphertext[:0], make([]byte, 12), plaintext, selfPub)
	l := len(ciphertext)
	ciphertext = append(ciphertext, selfPub...)
	if l > 0xff {
		return nil, errors.New("error of ecc.Encrypt")
	}
	ciphertext = append(ciphertext, byte(l))

	return ciphertext, nil
}

// Decrypt 私钥解密
func Decrypt(prikey []byte, ciphertext []byte) (plaintext []byte, err error) {

	l, ctLen := len(ciphertext), ciphertext[len(ciphertext)-1]
	if int(ctLen)+16 >= l { // 加密明文为空时取等号
		return nil, errors.New("ciphertext format error of ecc.Decrypt")
	}

	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), ciphertext[ctLen:l-1])
	if ok := (elliptic.P256()).IsOnCurve(x, y); !ok {
		return nil, errors.New("parse ECC publice key with ELLIPTIC(ANSI X9.62) error")
	}
	var key *ecdsa.PrivateKey
	if key, err = x509.ParseECPrivateKey(prikey); err != nil {
		plaintext = nil
		return
	}
	x, _ = key.Curve.ScalarMult(x, y, key.D.Bytes())
	if x == nil {
		return nil, errors.New("decrypt fail")
	}
	shard := sha256.Sum256(x.Bytes())

	// 解密
	var gcm cipher.AEAD
	if block, err := aes.NewCipher(shard[:]); err != nil {
		return nil, err
	} else {
		if gcm, err = cipher.NewGCM(block); err != nil {
			return nil, err
		}
	}

	plaintext = make([]byte, 0, len(ciphertext))
	plaintext, err = gcm.Open(plaintext[:0], make([]byte, 12), ciphertext[:ctLen], ciphertext[ctLen:l-1])
	if err != nil {
		plaintext = nil
	}
	return plaintext[:], err
}

// Sign 私钥加密
// 	签名的结构为：{x y x的长度(1B)}
func Sign(prikey []byte, data []byte) (signature []byte, err error) {
	var key *ecdsa.PrivateKey
	if key, err = x509.ParseECPrivateKey(prikey); err != nil {
		return nil, err
	}

	r, s, err := ecdsa.Sign(rand.Reader, key, data)
	if err != nil {
		return nil, err
	} else if r == nil || s == nil {
		return nil, errors.New("sign fail")
	}

	var rb, sb []byte
	if rb, err = r.MarshalText(); err != nil {
		return nil, err
	}
	if sb, err = s.MarshalText(); err != nil {
		return nil, err
	}

	// 最后字节的值表示rb的长度
	var res []byte = make([]byte, 0, len(rb)+len(sb)+1)
	res = append(res, rb...)
	res = append(res, sb...)
	if len(rb) > 0xff {
		return nil, errors.New("sign fail")
	}
	res = append(res, uint8(len(rb)))

	return res, nil
}

// Verify 公钥验签
func Verify(pubkey []byte, signature []byte, data []byte) (bool, error) {

	var rint, sint big.Int

	var l = len(signature)
	if l > 0xff || signature[l-1] > uint8(l) {
		return false, errors.New("verify fail: invalid format of signature")
	}
	if err := rint.UnmarshalText(signature[:signature[l-1]]); err != nil {
		return false, err
	}
	if err := sint.UnmarshalText(signature[signature[l-1] : l-1]); err != nil {
		return false, err
	}

	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), pubkey)
	if x == nil || y == nil {
		return false, errors.New("varify fail")
	} else {
		if ok := elliptic.P256().IsOnCurve(x, y); !ok {
			return false, errors.New("varify fail")
		}
	}
	var sps ecdsa.PublicKey = ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}

	// 根据公钥，明文，r，s验证签名
	return ecdsa.Verify(&sps, data, &rint, &sint), nil
}
