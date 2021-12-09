package cert

// 证书相关  仅支持圆锥曲线256

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// ParseCertificate 解析证书, 支持PEM、DER格式
func ParseCertificate(cert []byte) (*x509.Certificate, error) {
	var d []byte
	if p, _ := pem.Decode(cert); p != nil {
		d = p.Bytes
	} else {
		d = cert
	}
	return x509.ParseCertificate(d)
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

// GetCertInfo 提取ECC证书中的公钥、签名、数据
func GetCertInfo(cert []byte) (PublicKey []byte, Signature []byte, data []byte, err error) {
	return
}

func CreateEccCert(caCert *x509.Certificate, rootCert ...*x509.Certificate) (cert []byte, key []byte, err error) {
	return
}

func CertFormatCheck(cert []byte) bool {
	_, _, _, err := GetCertInfo(cert)
	return err == nil
}
