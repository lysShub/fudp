package cert

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/lysShub/fudp/internal/crypter/ecc"
)

// 证书相关

var temp *x509.Certificate = &x509.Certificate{
	SerialNumber: big.NewInt(int64(0)), // CA颁发证书对应的唯一序列号，自签填个随机数即可
	NotBefore:    time.Now(),
	NotAfter:     time.Now().AddDate(0, 0, 1), // 有效期结束时间

	KeyUsage:              x509.KeyUsageDigitalSignature,
	ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	BasicConstraintsValid: true,

	// 按需添加信息
	// IPAddresses: nil,
}

// CreateEccCert 生成ECC 256证书
//	rootCert为nil则为自签证书
// 参考: https://golang.org/src/crypto/tls/generate_cert.go
func CreateEccCert(rootCert *x509.Certificate, tmplate ...*x509.Certificate) (cert []byte, key []byte, err error) {

	privatekey, _, err := ecc.GenerateKey()
	if err != nil {
		return nil, nil, err
	}

	// 模板
	var myTemp, myRootTemp *x509.Certificate = temp, temp
	if len(tmplate) != 0 {
		myTemp = tmplate[0]
	}
	if rootCert != nil {
		myRootTemp = rootCert
	}

	var pri interface{} // *ecdsa.PrivateKey
	if pri, err = x509.ParseECPrivateKey(privatekey); err != nil {
		return nil, nil, err
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, myTemp, myRootTemp, &(pri.(*ecdsa.PrivateKey).PublicKey), pri)
	if err != nil {
		panic(err)
	}

	var certBytes = new(bytes.Buffer)
	if err := pem.Encode(certBytes, &pem.Block{Type: "ECC256 CERTIFICATE", Bytes: derBytes}); err != nil {
		panic(err)
	}

	var keyBytes = new(bytes.Buffer)
	if err := pem.Encode(keyBytes, &pem.Block{Type: "PRIVATE KEY", Bytes: privatekey}); err != nil {
		panic(err)
	}

	return certBytes.Bytes(), keyBytes.Bytes(), nil
}

// VerifyCertificate 校验证书
// 	rootCertificatePEM 为空则使用系统自带根证书校验
// 	返回nil表示校验成功
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

// CertFormatCheck 校验证书格式
func CertFormatCheck(cert []byte) bool {
	_, _, _, err := GetCertInfo(cert)
	return err == nil
}

// GetCertInfo 提取ECC证书中的公钥、签名、数据
func GetCertInfo(cert []byte) (PublicKey []byte, Signature []byte, data []byte, err error) {
	p, _ := pem.Decode(cert)
	if p == nil {
		panic("证书不是PEM格式")
	}

	r, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		panic(err)
	}

	if pub, ok := (r.PublicKey).(*ecdsa.PublicKey); ok {

		return elliptic.MarshalCompressed(elliptic.P256(), pub.X, pub.Y), r.Signature, r.Raw, nil

	} else {
		return nil, nil, nil, errors.New("invalid type of publice key, Type: " + fmt.Sprintf("%T", pub))
	}
}

// GetKeyInfo 获取ECC密钥的信息
func GetKeyInfo(key []byte) (prikey []byte, err error) {
	p, r := pem.Decode(key)
	if len(r) != 0 {
		return nil, errors.New("key parse fail")
	}
	return p.Bytes, nil
}
