package main

import (
	"flag"
	"fmt"
	"net"
	_ "net/http/pprof"

	"github.com/lysShub/fudp"
)

func main() {
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

		if err = f.HandReceive(nil); err != nil {
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

		if err = f.HandSend(""); err != nil {
			panic(err)
		}
	}

	return

}
