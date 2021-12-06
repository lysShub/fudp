package cert

import (
	"net"
	"testing"
	"time"
)

func TestCert(t *testing.T) {

	ca, key, err := GenerateCert(time.Now().AddDate(1, 0, 0).Sub(time.Now()), func(c *CaRequest) {
		c.AddDNSNames("getwenku.com").AddIPAddresses(net.ParseIP("4.4.4.4")).AddEmailAddresses("admin@getwenku.com").AddSubject("SERIALNUMBER=123&OU=组织内单位1&O=组织1|组织2&POSTALCODE=邮编&STREET=街道&L=地区&ST=省&C=国家")
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ca)
	t.Log(key)

}

func TestMain() {
	// u, err := url.Parse("getnwenku.com")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// ca, key, err := GenerateCert(time.Now().AddDate(1, 0, 0).Sub(time.Now()), func(c *cert.CaRequest) {
	// 	c.AddDNSNames("getwenku.com").AddIPAddresses(net.ParseIP("4.4.4.4")).AddEmailAddresses("admin@getwenku.com").AddURIs(u)
	// 	c.AddSubject("SERIALNUMBER=123&CN=通用名称&OU=组织内单位1&O=组织1|组织2&POSTALCODE=邮编&STREET=街道&L=地区&ST=省&C=国家")
	// })
	// if err != nil {
	// 	panic(err)
	// }

	// fh, _ := os.Create("ca.der")
	// fh.Write(ca)
	// fh.Close()

	// fh, _ = os.Create("key.pem")
	// fh.Write(key)
	// fh.Close()
	// csr, key, err := GenerateCsr(func(c *CaRequest) {
	// 	c.AddDNSNames("getwenku.com").AddIPAddresses(net.ParseIP("4.4.4.4")).AddEmailAddresses("admin@getwenku.com").AddURIs(u)
	// 	c.AddSubject("SERIALNUMBER=123&CN=通用名称&OU=组织内单位1&O=组织1|组织2&POSTALCODE=邮编&STREET=街道&L=地区&ST=省&C=国家")
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// fh, err = os.Create("csr.pem")
	// fh.Write(csr)
	// fh.Close()
}
