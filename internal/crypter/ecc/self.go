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

// GenerateKey
func GenerateKey() (priKey, pubKey []byte, err error) {
	var key *ecdsa.PrivateKey
	if key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err != nil {
		priKey, pubKey = nil, nil
	} else {

		if priKey, err = x509.MarshalPKCS8PrivateKey(key); err != nil {
			return nil, nil, err
		}

		pubKey = elliptic.MarshalCompressed(elliptic.P256(), key.X, key.Y)
	}
	return
}

// Encrypt 公钥加密
// 	密文结构为：{加密后密文 公钥 公钥长度(1B,单位字节)}
func Encrypt(pubkey []byte, plaintext []byte) (ciphertext []byte, err error) {

	var selfKey *ecdsa.PrivateKey
	if selfKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err != nil {
		return nil, err
	}
	selfPub := elliptic.MarshalCompressed(elliptic.P256(), selfKey.PublicKey.X, selfKey.PublicKey.Y)

	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), pubkey)
	if x == nil || y == nil {
		return nil, errors.New("invalid public key")
	} else {
		if ok := elliptic.P256().IsOnCurve(x, y); !ok {
			return nil, errors.New("invalid public key")
		}
	}
	var sps ecdsa.PublicKey = ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
	sx, sy := sps.Curve.ScalarMult(x, y, selfKey.D.Bytes())
	if sx == nil || sy == nil {
		return nil, errors.New("encrypt fail")
	}
	h := sha256.New()
	h.Write(sx.Bytes())
	h.Write(sy.Bytes())
	shard := h.Sum(nil) // AES_GCM_256对称加密密钥

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
	ciphertext = append(ciphertext, selfPub...)
	spkl := len(selfPub)
	if spkl > 0xff {
		return nil, errors.New("self public key too long")
	}
	ciphertext = append(ciphertext, byte(spkl))

	return ciphertext, nil
}

// Decrypt 私钥解密
func Decrypt(prikey []byte, ciphertext []byte) (plaintext []byte, err error) {

	l, pkl := len(ciphertext), int(ciphertext[len(ciphertext)-1])
	if pkl+1 >= l { // 密文为空时取等号
		return nil, errors.New("invalid ciphertext, wrong format")
	}

	ct, pk := ciphertext[0:l-pkl-1], ciphertext[l-pkl-1:l-1]

	sx, sy := elliptic.UnmarshalCompressed(elliptic.P256(), pk)
	if sx == nil || sy == nil {
		return nil, errors.New("invalid ciphertext")
	} else if ok := (elliptic.P256()).IsOnCurve(sx, sy); !ok {
		return nil, errors.New("invalid ciphertext")
	}

	var key *ecdsa.PrivateKey
	if tkey, err := x509.ParsePKCS8PrivateKey(prikey); err != nil {
		return nil, err
	} else {
		var ok bool = false
		if key, ok = tkey.(*ecdsa.PrivateKey); !ok {
			return nil, errors.New("invalid private key")
		}
	}

	sx, sy = key.Curve.ScalarMult(sx, sy, key.D.Bytes())
	if sx == nil {
		return nil, errors.New("decrypt fail")
	}
	h := sha256.New()
	h.Write(sx.Bytes())
	h.Write(sy.Bytes())
	shard := h.Sum(nil)

	// 解密
	var gcm cipher.AEAD
	if block, err := aes.NewCipher(shard[:]); err != nil {
		return nil, err
	} else {
		if gcm, err = cipher.NewGCM(block); err != nil {
			return nil, err
		}
	}

	plaintext, err = gcm.Open(nil, make([]byte, 12), ct, pk)
	if err != nil {
		plaintext = nil
	}
	return plaintext[:], err
}

// Sign 私钥加密
// 	签名的结构为：{x y x的长度(1B)}
func Sign(prikey []byte, data []byte) (signature []byte, err error) {
	var key *ecdsa.PrivateKey
	if tkey, err := x509.ParsePKCS8PrivateKey(prikey); err != nil {
		return nil, err
	} else {
		var ok bool = false
		if key, ok = tkey.(*ecdsa.PrivateKey); !ok {
			return nil, errors.New("invalid private key")
		}
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
		return false, errors.New("invalid signature, wrong  format")
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
