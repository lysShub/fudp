package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"go.uber.org/atomic"
)

var mtu int = 1300
var sAddr = net.UDPAddr{IP: net.ParseIP(""), Port: 19986}
var c atomic.Uint64
var blockSize int = 64

func main() {

}

//

func client() {
	go recorder()

	conn, err := net.DialUDP("udp", nil, &sAddr)
	if err != nil {
		panic(err)
	}

	var da = make([]byte, mtu)
	rand.Read(da)
	go func() {
		for i := 0; ; i++ {
			// 发送
			copy(da[0:], (*(*[8]byte)(unsafe.Pointer(&i)))[:])
			if _, err := conn.Write(da); err != nil {
				panic(err)
			}
			c.Add(1)
		}
	}()

	var rda = make([]byte, mtu)
	var t [8]byte
	var cc int
	if n, err := conn.Read(rda); err != nil {
		panic(err)
	} else if n >= 2 {
		// 第一字节是0表示echo, 是1表示重发
		if rda[0] == 0 {
			t = [8]byte{}
			copy(t[0:], rda[1:n])
			cc = *(*int)(unsafe.Pointer(&t))
			fmt.Println(cc)
		} else if rda[0] == 1 {
			for j := 1; j < n-8; j = j + 8 {
				copy(da[0:], (*(*[8]byte)(unsafe.Pointer(&j)))[:])
				if _, err := conn.Write(da); err != nil {
					panic(err)
				}
				c.Add(1)
			}
		}
	}
}

func Server() {
	var rda = make([]byte, mtu)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: sAddr.Port})
	if err != nil {
		panic(err)
	}

	var c atomic.Int64 = atomic.Int64{}
	var t [8]byte = [8]byte{}
	for {
		if n, raddr, err := conn.ReadFromUDP(rda); err == nil {
			panic(err)
		} else if n >= 8 {
			c.Add(1)
			copy(t[:], rda[0:])
			in := *(*int)(unsafe.Pointer(&t))
			fmt.Println(in, raddr)
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

func recorder() {
	for {
		time.Sleep(time.Second)
		s := c.Load() * uint64(mtu)
		fmt.Printf("%s \r", formatMemused(s))

		logHandle.Write(append([]byte(strconv.Itoa(int(s))), 10))
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

type spin struct {
	// put、get操作应该是事务性的，但现在不是；
	// 因为结合当前业务, 即使错过了阻塞释放，立马还有下一次机会
	block chan struct{}
	len   atomic.Int64
}

func NewSpin() *spin {
	return &spin{
		block: make(chan struct{}),
		len:   atomic.Int64{},
	}
}

func (s *spin) get() {
	if s.len.CAS(0, 0) {
		<-s.block
	} else {
		s.len.Add(-1)
	}
}

func (s *spin) put(n int) {
	if s.len.CAS(0, 0) {
		s.block <- struct{}{}
		s.len.Add(int64(n - 1))
	} else {
		s.len.Add(int64(n))
	}
}

type loss struct {
	m map[int]struct{}
	// r    []int
	maxi atomic.Uint64
	sync.RWMutex
}

func NewLoss() *loss {
	return &loss{
		m: make(map[int]struct{}, 64),
		// r:    make([]int, 0, 128),

		maxi: *atomic.NewUint64(1<<64 - 1),
	}
}

func (l *loss) put(n int) {
	if l.maxi.CAS(uint64(n)-1, uint64(n)) {
		return
	} else {
		if uint64(n) > l.maxi.Load() { // append
			l.Lock()
			for i := int(l.maxi.Load()) + 1; i < n; i++ {
				l.m[i] = struct{}{}
			}
		} else {
			delete(l.m, n)
		}
	}
}

func (l *loss) loss() []byte {
	max := int(l.maxi.Load())

	for k, _ := range l.m {
		if k+blockSize < max-1 {

		}
	}

	return nil
}

/*
  每个数据包都回声，收到一次回声后才能发送一个数据包
*/
