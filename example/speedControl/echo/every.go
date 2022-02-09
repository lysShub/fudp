package main

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
	client()
}

var mtu int = 1300
var sAddr = net.UDPAddr{IP: net.ParseIP("103.79.76.46"), Port: 19986}
var windowSize int = 1

func client() {
	go recorder()

	conn, err := net.DialUDP("udp", nil, &sAddr)
	if err != nil {
		panic(err)
	}

	var da = make([]byte, mtu)
	var rda = make([]byte, mtu)
	rand.Read(da)
	var cc atomic.Int64 = *atomic.NewInt64(0)

	go func() {
		sda := make([]byte, len(da))
		copy(sda, da)

		for {
			in := cc.Add(1) - 1
			copy(sda[0:], (*(*[8]byte)(unsafe.Pointer(&(in))))[:])

			if n, err := conn.Write(sda); err != nil {
				panic(err)
			} else {
				c.Add(uint64(n))
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()

	var tmp [8]byte = [8]byte{}
	for {
		for i := 0; i < windowSize; i++ {
			in := cc.Add(1) - 1
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

/*
  什么时候echo:
	当前WindowSize为n时，当读取了n个数据包时发送新的echo
	当遇到foo时发送新的echo

  foo情况:
	遇到乱序、丢包时, 获得的包序号与期望的包序号的差的绝对值为diff,
	1. 当发生一对顺序交换时, WindowSize减1, 当发生n个顺序交换时WindowSize减2n; 但是不立即echo，把diff累计起来, 直到累计值达到5%当前WindowSize时才echo


*/
func server() {
	var rda = make([]byte, mtu)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: sAddr.Port})
	if err != nil {
		panic(err)
	}
	var imax atomic.Int64 = *atomic.NewInt64(0)
	var windowSize int = 1
	var diff int
	var tmp [8]byte = [8]byte{}
	for {
		var n int
		var raddr *net.UDPAddr
		var err error
		for i := 0; i < windowSize; i++ {
			if n, raddr, err = conn.ReadFromUDP(rda); err != nil {
				panic(err)
			} else if n >= 8 {

				copy(tmp[:], rda[0:])
				i := *(*int)(unsafe.Pointer(&tmp))
				if i-int(imax.Load()) > 1 {
					imax.Store(int64(i))
					diff += (i - windowSize) * 2
					if diff > 16 || diff*10 > windowSize {
						windowSize -= diff
						diff = 0
						goto echo // echo
					}
				}
				imax.Store(int64(i))

			}
		}

		windowSize += 1
	echo:
		tmp = *(*[8]byte)(unsafe.Pointer(&windowSize))
		if _, err := conn.WriteToUDP(tmp[:], raddr); err != nil {
			panic(err)
		}

		logHandle.Write(append([]byte(strconv.Itoa(int(windowSize))), 10))

	}
}

var c atomic.Uint64

func recorder() {
	for {
		time.Sleep(time.Second)
		s := c.Load()
		fmt.Printf("%s/s  %v \r", formatMemused(s), windowSize)
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

var logHandle *os.File

func init() {
	var err error
	logHandle, err = os.OpenFile("./speed.log", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
}
