package main

import (
	"fmt"
	"math/rand"
	"net"
	"time"
	"unsafe"

	"go.uber.org/atomic"
)

/*
 接收方周期性echo, 通过设置发送方的的windowSize调整发送速度
 为了windowSize在合理范围内, 可以动态调整echo周期来确保流控的颗粒度
*/

var mtu int = 1300
var sAddr = net.UDPAddr{IP: net.ParseIP("103.79.76.46"), Port: 19986}
var initWindowSize int = 8

func client() {
	var windowSize int = 8 // 初始window
	go recorder(&windowSize)

	conn, err := net.DialUDP("udp", nil, &sAddr)
	if err != nil {
		panic(err)
	}

	var da = make([]byte, mtu)
	var rda = make([]byte, mtu)
	rand.Read(da)
	var secq atomic.Int64 = *atomic.NewInt64(0)
	var speed *atomic.Int64 = atomic.NewInt64(0)

	var tmp [8]byte = [8]byte{}
	for {
		for i := 0; i < windowSize; i++ {
			in := secq.Add(1) - 1
			copy(da[0:], (*(*[8]byte)(unsafe.Pointer(&(in))))[:])
			if n, err := conn.Write(da); err != nil {
				panic(err)
			} else {
				c.Add(uint64(n))
			}
		}

		if n, err := conn.Read(rda); err != nil {
			panic(err)
		} else if n >= 8 {
			copy(tmp[0:], rda[0:])
			windowSize = *(*int)(unsafe.Pointer(&tmp))
		}
	}
}

func server() {

	var rda = make([]byte, mtu)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: sAddr.Port})
	if err != nil {
		panic(err)
	}

	// var windowSize int = initWindowSize
	var loss *atomic.Int64 = atomic.NewInt64(0)
	var arvi *atomic.Int64 = atomic.NewInt64(0)
	var imax *atomic.Int64 = atomic.NewInt64(0)
	var echod time.Duration = time.Millisecond * 20
	var speed *atomic.Int64 = atomic.NewInt64(0)
	var lowEchoMode *atomic.Bool = atomic.NewBool(true)
	var raddr *net.UDPAddr

	// 流控, 
	go func() {
		var tmp int64
		var da = make([]byte, 8)
		for {
			time.Sleep(echod)

			if raddr == nil || !lowEchoMode.Load() {
				continue
			}
			if loss.Load() > 0 {
				tmp = newWindowSize(arvi, loss)
				if tmp > 10 {

				}
			} else {
				tmp = arvi.Load() + 1
			}
			n := copy(da[0:], (*(*[8]byte)(unsafe.Pointer(&tmp)))[:])
			conn.WriteToUDP(da[:n], raddr)

			loss.Store(0)
			arvi.Store(0)
		}
	}()

	// echo 模式转换
	go func() {
		var boundary = 
		for {
			time.Sleep(time.Millisecond * 100)

		}
	}()

	var tmp [8]byte = [8]byte{}
	var n int
	for {
		if n, raddr, err = conn.ReadFromUDP(rda); err != nil {
			panic(err)
		} else if n > 8 {
			copy(tmp[:], rda[0:])
			i := *(*int64)(unsafe.Pointer(&tmp))

			if d := i - imax.Load(); d > 0 {
				imax.Store(int64(i))
				if d > 1 {
					loss.Add(d * 2)
				}
			} else {
				arvi.Add(1)
			}
			speed.Add(1)
		}
	}
}

var c atomic.Uint64

func recorder(w *int) {
	for {
		time.Sleep(time.Second)
		s := c.Load()
		fmt.Printf("%s/s  %v \r", formatMemused(s), *w)
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

// loss 已经不为0了
// new WindowSize 始终不为0
func newWindowSize(arvi, loss *atomic.Int64) (r int64) {
	if loss.Load()*7 > arvi.Load() {
		r = arvi.Load() / 2
	} else {
		r = arvi.Load() - loss.Load()*2
	}

	if r < 1 {
		r = 1
	}
	return
}
