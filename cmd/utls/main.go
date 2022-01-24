package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/lysShub/fudp/internal/ioer"
)

func handleConn(conn net.Conn) {
	defer conn.Close()

	var buf []byte = make([]byte, 2000)
	if _, err := conn.Write([]byte("server")); err != nil {
		panic(err)
	}
	fmt.Println("进入")

	for {
		if n, err := conn.Read(buf); err != nil && err != io.EOF {
			panic(err)
		} else {
			fmt.Println("readed")
			conn.Write(buf[:n])
			fmt.Println("writed")
		}
	}
}

func server() {

	// D:/OneDrive/code/go/ctest/
	cert, err := tls.LoadX509KeyPair("D:/OneDrive/code/go/ctest//serve.cert.pem", "D:/OneDrive/code/go/ctest//serve.key.pem")
	if err != nil {
		panic(err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}

	l, err := ioer.Listen("udp", &net.UDPAddr{Port: 56635})
	if err != nil {
		panic(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		} else {
			fmt.Println("accetped!")
			tconn := tls.Server(conn, config)
			fmt.Println("handed!")
			handleConn(tconn)
		}
	}

	return
	// --------------------------------------------------------- //

	ln, err := tls.Listen("udp", "localhost:56635", config)
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		go handleConn(conn)
	}
}

func main() {
	go server()

	time.Sleep(time.Second)
	client()

}

func A(conn net.Conn) {
	fmt.Println(conn)
}

func client() {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("udp", "127.0.0.1:56635", conf)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Println("client connected")
	return
	_, err = conn.Write([]byte("hello"))
	if err != nil {
		panic(err)
	}
	buf := make([]byte, 100)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buf[:n]))
}
