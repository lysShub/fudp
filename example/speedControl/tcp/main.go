package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"go.uber.org/atomic"
)

// 测试tcp的传输曲线

func main() {
	if len(os.Args) == 1 {
		server()
	} else {
		client()
	}
}

var mtu int = 1300
var sAddr = net.TCPAddr{IP: net.ParseIP("172.23.128.92"), Port: 19986}

var speed = atomic.NewInt64(0)

func client() {
	conn, err := net.DialTCP("tcp", nil, &sAddr)
	if err != nil {
		panic(err)
	}
	var da = make([]byte, mtu)
	rand.Read(da)
	go recorder()

	for {
		if n, err := conn.Write(da); err != nil {
			panic(err)
		} else {
			speed.Add(int64(n))
		}
	}

}

func server() {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{Port: sAddr.Port})
	if err != nil {
		panic(err)
	}

	for {
		fmt.Println("sever accept")
		conn, err := l.AcceptTCP()
		if err != nil {
			panic(err)
		}
		go func() {
			var da = make([]byte, mtu)
			for {
				conn.Read(da)
			}
		}()
	}
}

var logHandle *os.File

func init() {
	var err error
	logHandle, err = os.OpenFile("./speed.txt", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
}

func recorder() {
	for {
		time.Sleep(time.Second)
		logHandle.Write(append([]byte(strconv.Itoa(int(speed.Load()))), 10))
		fmt.Printf("%v/s \r", formatMemused(uint64(speed.Load())))
		speed.Store(0)
	}
}

func formatMemused(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "KMGTPE"[exp])
}
