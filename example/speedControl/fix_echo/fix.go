package main

/*
  接收方固定20ms echo, 通过调整WindowSize来控制速度
*/

import (
	"fmt"
	"math/rand"
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
var sAddr = net.UDPAddr{IP: net.ParseIP("172.28.249.120"), Port: 19986}

// var sAddr = net.UDPAddr{IP: net.ParseIP("114.116.254.26"), Port: 19986}

var logHandle *os.File

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
	var ispeed *atomic.Uint64 = atomic.NewUint64(0)
	var ch chan int = make(chan int, 1)
	ch <- 8 // 初始

	// 获取windowSize
	go func() {
		var windowSize int
		go func() { // 速度记录器
			for {
				time.Sleep(time.Second)
				s := ispeed.Load()
				logHandle.Write(append([]byte(strconv.Itoa(int(s))), 10))
				fmt.Printf("%v/s \r", formatMemused(s))
				ispeed.Store(0)
			}
		}()

		var rda = make([]byte, 2000)
		var tmp [8]byte = [8]byte{}
		for {
			if n, err := conn.Read(rda); err != nil {
				panic(err)
			} else if n >= 8 {
				copy(tmp[0:], rda[0:])
				windowSize = *(*int)(unsafe.Pointer(&tmp))
				select {
				case ch <- windowSize:
				default:
				}
			} else {
				panic("")
			}
		}
	}()

	// 发送
	var ws int
	var da = make([]byte, mtu)
	rand.Read(da)
	for {
		ws = <-ch

		for i := 0; i < ws; i++ {
			in := secq.Add(1) - 1
			copy(da[0:], (*(*[8]byte)(unsafe.Pointer(&(in))))[:])
			if n, err := conn.Write(da); err != nil {
				panic(err)
			} else {
				ispeed.Add(uint64(n))
			}
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
	var qps *atomic.Int64 = atomic.NewInt64(0)
	var iqps *atomic.Int64 = atomic.NewInt64(0)
	go func() {
		var w time.Duration = time.Millisecond*31 + 250*time.Microsecond
		for {
			time.Sleep(w)
			qps.Store(iqps.Load() << 5)
			iqps.Store(0)
		}
	}()

	var lc, ac int
	go func() {
		for {
			time.Sleep(time.Second * 10)
			fmt.Println(lc, ac)
		}
	}()

	var raddr *net.UDPAddr
	// echo 流控
	go func() {

		var d time.Duration = time.Millisecond * 20
		var newWindowSize int64
		var r = ring{}
		for {

			if loss.Load() > 0 {
				// if loss.Load() < arvi.Load()/16 {
				// 	newWindowSize = arvi.Load() * 15 / 16
				// } else {
				// 	newWindowSize = arvi.Load() - loss.Load()
				// 	fmt.Println(arvi.Load(), loss.Load(), arvi.Load()/loss.Load())
				// }
				// newWindowSize = arvi.Load() - loss.Load()*2
				newWindowSize = r.avg(256) * 2 / 3
				lc++

			} else {
				newWindowSize = arvi.Load() + 1
				r.Put(newWindowSize)
				ac++
			}

			loss.Store(0)
			arvi.Store(0)
			if raddr != nil {
				if newWindowSize < 1 {
					newWindowSize = 1
				}
				conn.WriteToUDP((*(*[8]byte)(unsafe.Pointer(&newWindowSize)))[:], raddr)
			}
			time.Sleep(d)
		}
	}()

	var tmp [8]byte = [8]byte{}
	var n int
	for {
		if n, raddr, err = conn.ReadFromUDP(rda); err != nil {
			panic(err)
		} else if n > 8 {

			iqps.Add(1)
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
