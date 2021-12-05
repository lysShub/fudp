package cert

import "crypto/x509"

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

func VerifyCertificate(certificatePEM []byte, rootCertificatePEM ...[]byte) error {
	return nil
}
func GetCertPubkey(cert []byte) (pubkey []byte, err error) {
	return
}
