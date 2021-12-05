package main

import (
	"flag"
	"fmt"
	"net"
	_ "net/http/pprof"
	"net/url"
	"os"
	"time"

	"github.com/lysShub/fudp"
	"github.com/lysShub/fudp/internal/crypter/cert"
)

func main0() {
	u, err := url.Parse("getnwenku.com")
	if err != nil {
		fmt.Println(err)
		return
	}

	ca, key, err := cert.GenerateCert(time.Now().AddDate(1, 0, 0).Sub(time.Now()), func(c *cert.CaRequest) {
		c.AddDNSNames("getwenku.com").AddIPAddresses(net.ParseIP("4.4.4.4")).AddEmailAddresses("admin@getwenku.com").AddURIs(u)
		c.AddSubject("SERIALNUMBER=123&CN=通用名称&OU=组织内单位1&O=组织1|组织2&POSTALCODE=邮编&STREET=街道&L=地区&ST=省&C=国家")
	})
	if err != nil {
		panic(err)
	}

	fh, _ := os.Create("ca.der")
	fh.Write(ca)
	fh.Close()

	fh, _ = os.Create("key.pem")
	fh.Write(key)
	fh.Close()

	csr, key, err := cert.GenerateCsr(func(c *cert.CaRequest) {
		c.AddDNSNames("getwenku.com").AddIPAddresses(net.ParseIP("4.4.4.4")).AddEmailAddresses("admin@getwenku.com").AddURIs(u)
		c.AddSubject("SERIALNUMBER=123&CN=通用名称&OU=组织内单位1&O=组织1|组织2&POSTALCODE=邮编&STREET=街道&L=地区&ST=省&C=国家")
	})
	if err != nil {
		panic(err)
	}
	fh, err = os.Create("csr.pem")
	fh.Write(csr)
	fh.Close()
}

func main1() {
	isSend := flag.Bool("send", false, "指定运行方式")
	flag.Parse()
	// client:19986   <------>    server:19987

	var conn *net.UDPConn
	var token string
	var err error
	if *isSend {
		fmt.Println("发送")

		conn, err = net.DialUDP("udp", &net.UDPAddr{IP: nil, Port: 19986}, &net.UDPAddr{IP: nil, Port: 19987})
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		var c, err = fudp.Configure(func(c *fudp.Config) {
			c.PPMode().Send(`D:\OneDrive\code\go\fudp\test\main.go`)
		})
		if err != nil {
			panic(err)
		}

		token = c.ShowToken()
		fmt.Println(token)

		f, err := fudp.Run(c, conn)
		if err != nil {
			panic(err)
		}

		if _, err = f.HandReceive(); err != nil {
			panic(err)
		}

	} else {
		fmt.Println("接收")

		conn, err = net.DialUDP("udp", &net.UDPAddr{IP: nil, Port: 19987}, &net.UDPAddr{IP: nil, Port: 19986})
		if err != nil {
			panic(err)
		}

		defer conn.Close()

		fmt.Println("输入token:")
		fmt.Scanln(&token)
		fmt.Println(token)

		var c, err = fudp.Configure(func(c *fudp.Config) {
			c.PPMode().Receive(`D:\OneDrive\code\go\fudp\test`, c.ParseToken(token))
		})
		if err != nil {
			panic(err)
		}

		f, err := fudp.Run(c, conn)
		if err != nil {
			panic(err)
		}

		if _, err = f.HandSend(""); err != nil {
			panic(err)
		}
	}

	return

}
