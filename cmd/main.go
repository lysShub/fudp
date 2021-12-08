package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	_ "net/http/pprof"

	"github.com/lysShub/fudp"
	"github.com/lysShub/fudp/log"
)

func main() {
	fmt.Println(fudp.PPMode, fudp.CRole)

	log.Log(errors.New("this is a log"))
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

		f, err := fudp.Run(c, conn)
		if err != nil {
			panic(err)
		}
		token = f.ShowToken()
		fmt.Println(token)

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
			c.PPMode().Receive(`D:\OneDrive\code\go\fudp\test`, token)
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
