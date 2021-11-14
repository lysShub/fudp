package crypter

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// 证书相关

// CreateSelfSignedCert_ECC 生成ECC 256自签证书
//
// 参考: https://golang.org/src/crypto/tls/generate_cert.go
func CreateSelfSignedCert_ECC(path string) error {
	if fi, err := os.Stat(path); err != nil {
		return err
	} else if !fi.IsDir() {
		return errors.New("floder path" + path + " not exit")
	}
	var CrtPath, KeyPath string = filepath.Join(path, "ca.crt"), filepath.Join(path, "ca.key")

	var priv interface{}
	var err error
	if priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err != nil {
		panic(err)
	}

	if pri, ok := priv.(*ecdsa.PrivateKey); ok {
		publicKey, err := x509.MarshalPKIXPublicKey(&pri.PublicKey)
		if err != nil {
			panic(err)
		}
		fmt.Println(publicKey)
	} else {
		panic("格式解析错误")
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(int64(0)), // CA颁发证书对应的唯一序列号，自签填个随机数即可
		Subject: pkix.Name{
			Organization: []string{"Acme Co"}, // 组织名
		},
		NotBefore: time.Now(),                  // 有效期起始时间
		NotAfter:  time.Now().AddDate(1, 0, 0), // 有效期结束时间

		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		// 按需添加信息
		// IPAddresses: nil,
	}

	// 自签
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &(priv.(*ecdsa.PrivateKey).PublicKey), priv)
	if err != nil {
		panic(err)
	}

	certOut, err := os.OpenFile(CrtPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "ECC256 CERTIFICATE", Bytes: derBytes}); err != nil {
		panic(err)
	}
	certOut.Close()

	keyOut, err := os.OpenFile(KeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		panic(err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		panic(err)
	}
	keyOut.Close()

	return nil
}

// VerifyCertificate 校验证书
// 	rootCertificatePEM 为空则使用系统自带根证书校验
// 	返回nil表示校验成功，否则代表存在问题
func VerifyCertificate(certificatePEM []byte, rootCertificatePEM ...[]byte) error {

	var roots *x509.CertPool = x509.NewCertPool()
	for i := 0; i < len(rootCertificatePEM); i++ {
		if ok := roots.AppendCertsFromPEM([]byte(rootCertificatePEM[i])); !ok {
			return errors.New("failed to parse root certificate：rootCertificatePEM[" + strconv.Itoa(i) + "]")
		}
	}

	var block *pem.Block
	if block, _ = pem.Decode([]byte(certificatePEM)); block == nil {
		return errors.New("failed to parse certificate PEM")
	}

	if cert, err := x509.ParseCertificate(block.Bytes); err != nil {
		return errors.New("failed to parse certificate x509: " + err.Error())
	} else {
		opts := x509.VerifyOptions{
			Roots: roots,
		}

		if _, err := cert.Verify(opts); err != nil {
			return errors.New("failed to verify certificate " + err.Error())
		}
	}
	return nil
}

// CertificateToPubkey 提取ECC证书中的公钥
func CertificateToPubkey(cert string) []byte {

	c, err := os.OpenFile(cert, os.O_RDONLY, 0600)
	if err != nil {
		panic(err)
	}
	cda, err := io.ReadAll(c)
	if err != nil {
		panic(err)
	}
	p, _ := pem.Decode(cda)
	if p == nil {
		panic("证书不是PEM格式")
	}

	fmt.Println("p.Type", p.Type)

	r, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		panic(err)
	}

	if pub, ok := (r.PublicKey).(*ecdsa.PublicKey); ok {
		publicKey, err := x509.MarshalPKIXPublicKey(pub)
		if err != nil {
			panic(err)
		}
		fmt.Println(publicKey)

	} else {
		panic("格式解析错误")
	}

	return nil
}
