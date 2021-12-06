package cert

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/lysShub/fudp/internal/crypter/ecc"
)

// 生成证书

// GenerateCert 生成ECDSAWithSHA256算法的der格式自签证书和pem格式的私钥
// 	@timeout 证书有效期
// 	@rootCert 根证书, 如果为空则为自签根证书
func GenerateCert(timeout time.Duration, fun func(c *CaRequest), rootCert ...*x509.Certificate) (cert []byte, prikey *ecdsa.PrivateKey, err error) {
	var c = new(CaRequest)
	c.signatureAlgorithm = x509.ECDSAWithSHA256
	fun(c)
	if c.err != nil {
		return nil, nil, c.err
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

// SetSubject 生成CSR文件, 用于申请CA证书
// 	@key pem格式的私钥
func GenerateCsr(fun func(c *CaRequest)) (csr []byte, key []byte, err error) {
	var c = new(CaRequest)
	c.signatureAlgorithm = x509.ECDSAWithSHA256
	fun(c)
	if c.err != nil {
		return nil, nil, c.err
	}

	var cr x509.CertificateRequest = x509.CertificateRequest{
		SignatureAlgorithm: c.signatureAlgorithm,
		Subject:            c.subject,
		DNSNames:           c.dNSNames,
		EmailAddresses:     c.emailAddresses,
		IPAddresses:        c.iPAddresses,
		URIs:               c.uRIs,
	}

	privatekey, err := ecc.GenerateKey()
	if err != nil {
		return nil, nil, err
	}

	csrbin, err := x509.CreateCertificateRequest(rand.Reader, &cr, privatekey)
	if err != nil {
		return nil, nil, err
	}

	var keyBytes = new(bytes.Buffer)
	if err = pem.Encode(keyBytes, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrbin}); err != nil {
		return nil, nil, err
	}
	csr = keyBytes.Bytes()

	priBytes, err := x509.MarshalPKCS8PrivateKey(privatekey)
	if err != nil {
		return nil, nil, err
	}
	keyBytes = new(bytes.Buffer)
	if err = pem.Encode(keyBytes, &pem.Block{Type: "PRIVATE KEY", Bytes: priBytes}); err != nil {
		return nil, nil, err
	}
	key = keyBytes.Bytes()

	return
}

type CaRequest struct {
	signatureAlgorithm x509.SignatureAlgorithm
	subject            pkix.Name
	dNSNames           []string
	emailAddresses     []string
	iPAddresses        []net.IP
	uRIs               []*url.URL
	extraExtensions    []pkix.Extension
	attributes         []pkix.AttributeTypeAndValue

	err error
}

// AddSubject 用于添加证书的主体信息
//  字符串遵循以下格式: SERIALNUMBER=123&CN=通用名称&OU=组织内单位1&O=组织1|组织2&POSTALCODE=邮编&STREET=街道&L=地区&ST=省&C=国家
func (c *CaRequest) AddSubject(subjcet string) *CaRequest {
	if len(subjcet) == 0 {
		return c
	}

	subjcet = strings.ToUpper(subjcet)
	us, err := url.ParseQuery(subjcet)
	if err != nil {
		c.err = errors.New("invalid subject, wrong format")
		return c
	}
	c.subject = pkix.Name{}

	c.subject.SerialNumber = strings.TrimSpace(us.Get("SERIALNUMBER"))
	if _, err = strconv.Atoi(c.subject.SerialNumber); err != nil {
		c.err = errors.New("invalid subject, wrong SerialNumber")
		return c
	}
	c.subject.CommonName = strings.TrimSpace(us.Get("CN"))
	for _, v := range strings.Split(us.Get("OU"), "|") {
		c.subject.OrganizationalUnit = append(c.subject.OrganizationalUnit, strings.TrimSpace(v))
	}
	for _, v := range strings.Split(us.Get("O"), "|") {
		c.subject.Organization = append(c.subject.Organization, strings.TrimSpace(v))
	}
	for _, v := range strings.Split(us.Get("POSTALCODE"), "|") {
		c.subject.PostalCode = append(c.subject.PostalCode, strings.TrimSpace(v))
	}
	for _, v := range strings.Split(us.Get("STREET"), "|") {
		c.subject.StreetAddress = append(c.subject.StreetAddress, strings.TrimSpace(v))
	}
	for _, v := range strings.Split(us.Get("L"), "|") {
		c.subject.Locality = append(c.subject.Locality, strings.TrimSpace(v))
	}
	for _, v := range strings.Split(us.Get("ST"), "|") {
		c.subject.Province = append(c.subject.Province, strings.TrimSpace(v))
	}
	for _, v := range strings.Split(us.Get("C"), "|") {
		c.subject.Country = append(c.subject.Country, strings.TrimSpace(v))
	}
	return c
}

// AddDNSName 证书域名
func (c *CaRequest) AddDNSNames(dnsNames ...string) *CaRequest {
	c.dNSNames = append(c.dNSNames, dnsNames...)
	return c
}
func (c *CaRequest) AddEmailAddresses(emails ...string) *CaRequest {
	c.emailAddresses = append(c.emailAddresses, emails...)
	return c
}

// AddIPAddresses 证书IP
func (c *CaRequest) AddIPAddresses(ips ...net.IP) *CaRequest {
	c.iPAddresses = append(c.iPAddresses, ips...)
	return c
}

func (c *CaRequest) AddURIs(urls ...*url.URL) *CaRequest {
	c.uRIs = append(c.uRIs, urls...)
	return c
}
