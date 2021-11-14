package main

import (
	"fmt"
	_ "net/http/pprof"

	"github.com/lysShub/fudp/internal/crypter/cert"
	"github.com/lysShub/fudp/internal/crypter/ecc"
)

func main() {
	c, k, err := cert.CreateEccCert(nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(c))
	// fmt.Println("")
	// fmt.Println(k)
	pub, _, _, _ := cert.GetCertInfo(c)
	fmt.Println(pub)
	return
	fmt.Println(k)

	pri, pub, err := ecc.GenerateKey()
	if err != nil {
		panic(err)
	}
	da := []byte("0123456789abcdef0123456789abcdef")
	// fmt.Println("主公钥长度", len(pub))

	ct, err := ecc.Encrypt(pub, da)
	if err != nil {
		panic(err)
	}
	// fmt.Println("整个密文长度", len(ct))
	fmt.Println(ct)

	pt, err := ecc.Decrypt(pri, ct)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(pt))

	// 证书中da既是摘要
	sign, err := ecc.Sign(pri, da)
	if err != nil {
		panic(err)
	}

	fmt.Println(ecc.Verify(pub, sign, da))
}
