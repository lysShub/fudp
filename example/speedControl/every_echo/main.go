package main

// 接收到每个数据包都echo
import (
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
	"unsafe"

	"go.uber.org/atomic"
)

func main() {

	if len(os.Args) != 1 {
		client()
	} else {
		server()
	}

}

var mtu int = 1300

// var sAddr = net.UDPAddr{IP: net.ParseIP("103.79.76.46"), Port: 19986}

// var sAddr = net.UDPAddr{IP: net.ParseIP("172.23.128.92"), Port: 19986}

var sAddr = net.UDPAddr{IP: net.ParseIP("114.116.254.26"), Port: 19986}

var logHandle *os.File
var ispeed *atomic.Uint64 = atomic.NewUint64(0)

func init() {
	var err error
	logHandle, err = os.OpenFile("./speed.txt", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
}

func client() {
	conn, err := net.DialUDP("udp", nil, &sAddr)
	if err != nil {
		panic(err)
	}

	var secq *atomic.Int64 = atomic.NewInt64(0) // 递增sequence

	go func() { // 速度记录器
		for {
			time.Sleep(time.Second)
			s := ispeed.Load()
			logHandle.Write(append([]byte(strconv.Itoa(int(s))), 10))
			fmt.Printf("%v/s \r", formatMemused(s))
			ispeed.Store(0)
		}
	}()

	// 发送
	var ws int = 16
	var da = make([]byte, mtu)
	var rda = make([]byte, mtu)
	var tmp [8]byte = [8]byte{}
	rand.Read(da)
	go func() {
		for {
			time.Sleep(time.Millisecond * 2)
			n, _ := conn.Write(da)
			ispeed.Add(uint64(n))

		}
	}()
	for {
		for i := 0; i < ws; i++ {
			in := secq.Add(1) - 1
			copy(da[0:], (*(*[8]byte)(unsafe.Pointer(&(in))))[:])
			if n, err := conn.Write(da); err != nil {
				panic(err)
			} else {
				ispeed.Add(uint64(n))
			}
		}

		if n, err := conn.Read(rda); err != nil {
			panic(err)
		} else if n >= 8 {
			copy(tmp[0:], rda[0:])
			ws = *(*int)(unsafe.Pointer(&tmp))
		}

	}
}

func server() {
	fmt.Println("server")

	var rda = make([]byte, mtu)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: sAddr.Port}) // IP: net.ParseIP("172.31.229.13"),
	if err != nil {
		panic(err)
	}

	// var windowSize int = initWindowSize
	var loss *atomic.Int64 = atomic.NewInt64(0)
	var arvi *atomic.Int64 = atomic.NewInt64(0)
	var imax *atomic.Int64 = atomic.NewInt64(0) // 最大sequence

	// 速度记录器 qps
	go func() { // 速度记录器
		for {
			time.Sleep(time.Second)
			s := ispeed.Load()
			logHandle.Write(append([]byte(strconv.Itoa(int(s))), 10))
			fmt.Printf("%v/s \r", formatMemused(s))
			ispeed.Store(0)
		}
	}()

	var raddr *net.UDPAddr
	var tmp [8]byte = [8]byte{}
	var n int
	var m = uint64(1)
	for {
		for i := uint64(0); i < m; i++ {
			if n, raddr, err = conn.ReadFromUDP(rda); err != nil {
				panic(err)
			} else if n > 8 {
				ispeed.Add(uint64(n))
				copy(tmp[:], rda[0:])
				i := *(*int64)(unsafe.Pointer(&tmp)) // get sequence

				if d := i - imax.Load(); d > 0 {
					imax.Store(int64(i))
					if d > 1 {
						loss.Add(d)
					} else {
						arvi.Add(1)
					}
				} else {
					arvi.Add(1)
				}
			}
		}

		// echo
		if _, err = conn.WriteToUDP((*(*[8]byte)(unsafe.Pointer(&m)))[:], raddr); err != nil {
			panic(err)
		}

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

type ring struct {
	r [256]int64
	i uint8
}

func (r *ring) Put(v int64) {
	r.r[r.i+1] = v
	r.i = r.i + 1
}

func (r *ring) avg(n int) int64 {
	var s, c int64
	if n > 256 {
		n = 256
	}

	for i, j := r.i, n; j > 0; i, j = i-1, j-1 {
		if r.r[i] == 0 {
			break
		}
		s += r.r[i]
		c += 1
	}
	if c == 0 {
		return 0
	} else {
		return s / c
	}
}
