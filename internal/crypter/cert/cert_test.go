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
