package ecc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"golang.org/x/crypto/curve25519"
)

// 实现ECC_P256实现的非对称加密
// go官方库crypto/elliptic和crypto/ecdsa没有实现开箱即用的接口, 需要进行二次封装
//
// 但是现在并未向RAS一样有通用的标准, 因此自己进行实现
// 也可以可以添加自己的实现

// 必须实现以下函数

/*
func GenerateKey() (priKey, pubKey []byte, err error)
func Encrypt(pubkey []byte, plaintext []byte) (ciphertext []byte, err error)
func Decrypt(prikey []byte, ciphertext []byte) (plaintext []byte, err error)
func Sign(prikey []byte, data []byte) (signature []byte, err error)
func Verify(pubkey []byte, signature []byte, data []byte) (bool, error)
*/

type PrivateKey ecdsa.PrivateKey
type PublicKey ecdsa.PublicKey

func GenerateKey() (priKey *PrivateKey, err error) {
	pri, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return (*PrivateKey)(pri), nil
}

func (pri *PrivateKey) Decrypt(ciphertext []byte) []byte {
	return nil
}

func Encrypt(pub *PublicKey, msg []byte) (ciphertext []byte, err error) {
	peerPub := pub
	pri, err := GenerateKey()
	if err != nil {
		return nil, err
	}
	// golang.org/x/crypto/curve25519
	// pri pub
	curve25519.X25519(nil, peerPub)

	return nil
}
