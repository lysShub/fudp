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

type PrivateKey = ecdsa.PrivateKey
type PublicKey = ecdsa.PublicKey

func GenerateKey() (priKey *PrivateKey, err error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func MarshalPubKey(pubKey *PublicKey) (publicKey []byte, err error) {
	if pubKey.X == nil || pubKey.Y == nil {
		return nil, errors.New("invalid ecc public key")
	}
	if !elliptic.P256().IsOnCurve(pubKey.X, pubKey.Y) {
		return nil, errors.New("invalid ecc public key")
	}
	return elliptic.MarshalCompressed(elliptic.P256(), pubKey.X, pubKey.Y), nil
}

func ParsePubKey(publicKey []byte) (pubKey *PublicKey, err error) {
	x, y := elliptic.Unmarshal(elliptic.P256(), publicKey)
	if x == nil || y == nil {
		return nil, errors.New("invalid ecc public key")
	}
	if !elliptic.P256().IsOnCurve(x, y) {
		return nil, errors.New("invalid ecc public key")
	}
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
}

// MarshalPrikey PKCS8
func MarshalPrikey(priKey *PrivateKey) (privateKey []byte, err error) {
	return x509.MarshalPKCS8PrivateKey(priKey)
}

func ParsePriKey(privateKey []byte) (priKey *PrivateKey, err error) {

	if t, err := x509.ParsePKCS8PrivateKey(privateKey); err != nil {
		return nil, err
	} else if condition, ok := t.(*PrivateKey); ok {
		return condition, nil
	} else {
		return nil, errors.New("invalid ecc private key")
	}
}

// Encrypt 公钥加密
// 	密文结构为：{加密后密文 公钥 公钥长度(1B,单位字节)}
func Encrypt(pubKey *PublicKey, plaintext []byte) (ciphertext []byte, err error) {
	var selfKey *PrivateKey
	if selfKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err != nil {
		return nil, err
	}
	selfPub := elliptic.MarshalCompressed(elliptic.P256(), selfKey.PublicKey.X, selfKey.PublicKey.Y)

	var sps ecdsa.PublicKey = ecdsa.PublicKey{Curve: elliptic.P256(), X: pubKey.X, Y: pubKey.Y}
	sx, sy := sps.Curve.ScalarMult(pubKey.X, pubKey.Y, selfKey.D.Bytes())
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
func Decrypt(priKey *PrivateKey, ciphertext []byte) (plaintext []byte, err error) {

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

	sx, sy = priKey.Curve.ScalarMult(sx, sy, priKey.D.Bytes())
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
func Sign(priKey *PrivateKey, data []byte) (signature []byte, err error) {
	r, s, err := ecdsa.Sign(rand.Reader, priKey, data)
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
func Verify(pubKey *PublicKey, signature []byte, data []byte) (bool, error) {

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

	if pubKey.X == nil || pubKey.Y == nil {
		return false, errors.New("invalid ecc public key")
	} else {
		if ok := elliptic.P256().IsOnCurve(pubKey.X, pubKey.Y); !ok {
			return false, errors.New("invalid ecc public key")
		}
	}
	var sps ecdsa.PublicKey = ecdsa.PublicKey{Curve: elliptic.P256(), X: pubKey.X, Y: pubKey.Y}

	// 根据公钥，明文，r，s验证签名
	return ecdsa.Verify(&sps, data, &rint, &sint), nil
}
