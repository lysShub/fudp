package main

// 测试kcp

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	kcp "github.com/xtaci/kcp-go/v5"
	"go.uber.org/atomic"
	"golang.org/x/crypto/pbkdf2"
)

var mtu int = 1300

// var sAddr = net.UDPAddr{IP: net.ParseIP("103.79.76.46"), Port: 19986}

var sAddr = net.UDPAddr{IP: net.ParseIP("172.20.24.2"), Port: 19986}

// var sAddr = net.UDPAddr{IP: net.ParseIP("114.116.254.26"), Port: 19986}

var ispeed *atomic.Uint64 = atomic.NewUint64(0)

func main() {
	if len(os.Args) != 1 {
		client()
	} else {
		server()
	}
}

var key = pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
var block kcp.BlockCrypt

func init() {
	block, _ = kcp.NewAESBlockCrypt(key)
}

func server() {
	fmt.Println("server")
	if listener, err := kcp.ListenWithOptions(":"+strconv.Itoa(sAddr.Port), block, 10, 3); err == nil {
		for {
			s, err := listener.AcceptKCP()
			if err != nil {
				panic(err)
			}
			go func(conn *kcp.UDPSession) {
				buf := make([]byte, 4096)
				for {
					_, err := conn.Read(buf)
					if err != nil {
						panic(err)
					} else {
						// fmt.Printf("%v \r", n)

						// fmt.Println(n)
					}
				}
			}(s)
		}
	} else {
		panic(err)
	}
}

func client() {
	fmt.Println("client")
	var da = make([]byte, mtu)
	rand.Read(da)

	go func() { // 速度记录器
		for {
			time.Sleep(time.Second)
			s := ispeed.Load()
			fmt.Printf("%v/s \r", formatMemused(s))
			ispeed.Store(0)
		}
	}()

	if sess, err := kcp.DialWithOptions(sAddr.String(), block, 10, 3); err == nil {
		for {
			if n, err := sess.Write([]byte(da)); err != nil {
				panic(err)
			} else {
				ispeed.Add(uint64(n))
			}
		}
	} else {
		panic(err)
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
