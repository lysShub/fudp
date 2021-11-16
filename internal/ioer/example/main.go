package main

import (
	"fmt"
	"net"

	"github.com/lysShub/fudp/internal/ioer"
)

func main() {

	i, err := ioer.New("udp", &net.UDPAddr{IP: net.ParseIP("192.168.1.8"), Port: 19986})
	if err != nil {
		panic(err)
	}

	for {
		conn := i.Accept()
		fmt.Println("new conn", conn.RemoteAddr())
		go Handle(conn)
	}

}

func Handle(conn *ioer.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(buf[:n]))
	}
}
