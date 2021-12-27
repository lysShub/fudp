package cert

// 证书相关  仅支持圆锥曲线256

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"math/big"
	"strconv"
	"time"

	"github.com/lysShub/fudp/internal/crypter/ecc"
)

// ParseCertificate 解析证书, 支持PEM、DER格式
func ParseCertificate(cert []byte) (*x509.Certificate, error) {
	var d []byte
	if p, _ := pem.Decode(cert); p != nil {
		d = p.Bytes
	} else {
		d = cert
	}
	if c, err := x509.ParseCertificate(d); err != nil {
		return nil, err
	} else {
		if err := c.CheckSignature(x509.ECDSAWithSHA256, c.Raw, c.Signature); err != nil {
			return nil, err
		} else {
			return c, nil
		}
	}
}

// VerifyCertificate 验签证书
// 	@certificatePEM 待验签证书
// 	@rootCertificatePEM 自签根证书, 为空则使用CA证书验签
func VerifyCertificate(certificate []byte, rootCertificatePEM ...[]byte) error {
	var roots *x509.CertPool = x509.NewCertPool()
	for _, p := range rootCertificatePEM {
		if c, err := ParseCertificate(p); err != nil {
			return err
		} else {
			roots.AddCert(c)
		}
	}

	if cert, err := ParseCertificate(certificate); err != nil {
		return err
	} else {
		opt := x509.VerifyOptions{
			Roots: roots,
		}
		_, err = cert.Verify(opt)
		return err
	}
}

// ErrInvalidCertificateType 证书加密类型不正确, 仅支持ECC
var ErrInvalidCertificateType = errors.New("invalid certificate encryption type")

// GetCertPubkey 从证书中提取公钥
func GetCertPubkey(cert []byte) (pubkey *ecdsa.PublicKey, err error) {
	if c, err := ParseCertificate(cert); err != nil {
		return nil, err
	} else {
		if p, ok := c.PublicKey.(*ecdsa.PublicKey); ok {
			return p, nil
		} else {
			return nil, ErrInvalidCertificateType
		}
	}
}

// GenerateCert 生成ECDSAWithSHA256算法的der格式自签证书和DER格式的私钥
// 	@timeout 证书有效期
// 	@rootCert 验签根证书, 如果为空则为根自签证书
func GenerateCert(timeout time.Duration, fun func(c *Csr), rootCert ...*x509.Certificate) (cert []byte, prikey *ecc.PrivateKey, err error) {
	var c = new(Csr)
	c.signatureAlgorithm = x509.ECDSAWithSHA256
	if fun != nil {
		if fun(c); c.err != nil {
			return nil, nil, c.err
		}
	}

	// 根据csr 生成模板
	var serialNumber *big.Int = big.NewInt(0)
	if c.subject.SerialNumber != "" {
		if s, err := strconv.Atoi(c.subject.SerialNumber); err != nil {
			return nil, nil, errors.New("invalid subject, wrong SerialNumber")
		} else {
			serialNumber = big.NewInt(int64(s))
		}
	}
	if timeout < 1e9 {
		return nil, nil, errors.New("invalid timeout, too sort")
	}

	var template *x509.Certificate = &x509.Certificate{
		SerialNumber: serialNumber, // CA颁发证书对应的唯一序列号，自签填个随机数即可
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(timeout), // 有效期结束时间

		SignatureAlgorithm: c.signatureAlgorithm, // 签名算法
		KeyUsage:           x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth | x509.ExtKeyUsageClientAuth},
		IsCA:               false,
		DNSNames:           c.dNSNames,
		EmailAddresses:     c.emailAddresses,
		IPAddresses:        c.iPAddresses,
		URIs:               c.uRIs,

		Issuer:  c.subject, // 证书颁发者
		Subject: c.subject, // 证书持有者

		BasicConstraintsValid: true,
	}

	var root *x509.Certificate = template
	if len(rootCert) > 0 {
		root = rootCert[0]
	}

	privatekey, err := ecc.GenerateKey()
	if err != nil {
		return nil, nil, err
	}
	cert, err = x509.CreateCertificate(rand.Reader, template, root, privatekey.PublicKey, privatekey)
	if err != nil {
		return nil, nil, err
	}
	return cert, prikey, nil
}

// CheckCertFormat 校验证书格式与加密算法
func CheckCertFormat(cert []byte) bool {
	_, err := ParseCertificate(cert)
	return err == nil
}
