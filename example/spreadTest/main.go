package main

import (
	"crypto/rand"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
	"unsafe"

	"go.uber.org/atomic"
)

// 测试分布

func main() {
	// client()
	fmt.Println(http.ListenAndServe(":19986", http.FileServer(http.Dir("/home/worker"))))
}

//
var mtu int = 1300 // udp 数据包mtu
var sAddr = net.UDPAddr{IP: net.ParseIP("103.79.76.46"), Port: 19986}
var speed int = 15 << 20 // 1MB/s

func client() {
	go recorder()

	conn, err := net.DialUDP("udp", nil, &sAddr)
	if err != nil {
		panic(err)
	}
	sn := time.Duration(1e9 * mtu / speed)

	t := time.NewTicker(sn)

	var da = make([]byte, mtu)
	rand.Read(da)
	for i := 0; ; i++ {
		// 发送

		copy(da[0:], (*(*[8]byte)(unsafe.Pointer(&i)))[:])
		if n, err := conn.Write(da); err != nil {
			panic(err)
		} else {
			c.Add(uint64(n))
		}
		<-t.C
	}
}

func server() {
	go recorder()
	println("server 启动")

	var rda = make([]byte, mtu)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: sAddr.Port})
	if err != nil {
		panic(err)
	}

	var t [8]byte = [8]byte{}
	for {
		if n, _, err := conn.ReadFromUDP(rda); err != nil {
			panic(err)
		} else if n >= 8 {
			c.Add(uint64(n))
			copy(t[:], rda[0:])
			in := *(*int)(unsafe.Pointer(&t))
			log(in)
		}
	}
}

var logHandle *os.File

func init() {
	var err error
	logHandle, err = os.OpenFile("./speed.log", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
}

var last int = 0

func log(n int) {
	logHandle.Write(append([]byte(strconv.Itoa(int(n-last))+"  "+strconv.Itoa(n)), 10))
	last = n
}

var c atomic.Uint64 // 速度记录器

func recorder() {
	for {
		time.Sleep(time.Second)
		s := c.Load()
		fmt.Printf("%s/s \r", formatMemused(s))

		c.Store(0)
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
