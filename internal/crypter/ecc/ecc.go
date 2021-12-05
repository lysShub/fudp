package ecc

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
