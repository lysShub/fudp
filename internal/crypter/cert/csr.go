package cert

import (
	"bytes"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/lysShub/fudp/internal/crypter/ecc"
)

// 生成证书

// SetSubject 生成CSR文件, 用于申请CA证书
// 	@key pem格式的私钥
func GenerateCSR(fun func(c *Csr)) (csr []byte, key []byte, err error) {
	var c = new(Csr)
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

// GenerateCsrWithPriviate
func GenerateCsrWithPriviate(privatekey *ecc.PrivateKey, fun func(c *Csr)) (csr []byte, err error) {
	var c = new(Csr)
	c.signatureAlgorithm = x509.ECDSAWithSHA256
	fun(c)
	if c.err != nil {
		return nil, c.err
	}

	var cr x509.CertificateRequest = x509.CertificateRequest{
		SignatureAlgorithm: c.signatureAlgorithm,
		Subject:            c.subject,
		DNSNames:           c.dNSNames,
		EmailAddresses:     c.emailAddresses,
		IPAddresses:        c.iPAddresses,
		URIs:               c.uRIs,
	}

	csrbin, err := x509.CreateCertificateRequest(rand.Reader, &cr, privatekey)
	if err != nil {
		return nil, err
	}
	var keyBytes = new(bytes.Buffer)
	if err = pem.Encode(keyBytes, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrbin}); err != nil {
		return nil, err
	}
	csr = keyBytes.Bytes()

	return
}

type Csr struct {
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
func (c *Csr) AddSubject(subjcet string) *Csr {
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
func (c *Csr) AddDNSNames(dnsNames ...string) *Csr {
	c.dNSNames = append(c.dNSNames, dnsNames...)
	return c
}
func (c *Csr) AddEmailAddresses(emails ...string) *Csr {
	c.emailAddresses = append(c.emailAddresses, emails...)
	return c
}

// AddIPAddresses 证书IP
func (c *Csr) AddIPAddresses(ips ...net.IP) *Csr {
	c.iPAddresses = append(c.iPAddresses, ips...)
	return c
}

// AddURIs
func (c *Csr) AddURIs(urls ...*url.URL) *Csr {
	c.uRIs = append(c.uRIs, urls...)
	return c
}
