package main

/*
	sender在发送一段数据后等待receiver的回复，收到回复后再继续传输
*/

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"go.uber.org/atomic"
)

var conCurrent int = 1       // 并发数
var blockSize int = 64 << 20 // 块大小
var mtu int = 1300           // udp 数据包mtu
var sAddr = net.UDPAddr{IP: net.ParseIP(""), Port: 19986}
var c atomic.Uint64 // 速度记录器

func main() {

}

//-------------------------------------------------------------//

/*
  数据包第一字节表示并发信道序号
  数据包第二、三字节表示本block中的数据包序号
*/

func Client() {
	go recorder()

	conn, err := net.DialUDP("udp", nil, &sAddr)
	if err != nil {
		panic(err)
	}
	for i := 0; i < conCurrent; i++ {
		go func(ci int) {
			var da = make([]byte, mtu)
			var rda = make([]byte, mtu)
			rand.Read(da)
			da[0] = byte(ci)

			for {
				// 发送
				for i, pi := mtu, 0; i <= blockSize; i, pi = i+mtu, pi+1 {
					da[1], da[2] = byte(i), byte(i>>8)
					if _, err := conn.Write(da); err != nil {
						panic(err)
					}
					c.Add(1)
				}

				// 读取echo
			re:
				if n, err := conn.Read(rda); err != nil {
					panic(err)
				} else if n > 0 {
					for j := 0; j < n-1; j = j + 2 {
						da[1], da[2] = rda[j], rda[j+1]
						if _, err := conn.Write(da); err != nil {
							panic(err)
						}
						c.Add(1)
					}
					goto re
				}
			}
		}(i)
	}
}

type ch struct {
	r chan int
	s chan []int
}

func Server() {

	var rda = make([]byte, mtu)
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: sAddr.Port})
	if err != nil {
		panic(err)
	}

	var cReco []ch = make([]ch, conCurrent)

	for i := 0; i < conCurrent; i++ {
		go func(index int) {
			cReco[index] = ch{
				make(chan int, 16), make(chan []int, 16),
			}
			var t = time.NewTimer(time.Millisecond * 100)

			for {
				select {
				case j := <-cReco[index].r:
					fmt.Println(j)
				case <-t.C:
				}
				t.Reset(time.Microsecond * 100)
			}
		}(i)
	}

	for {
		if n, raddr, err := conn.ReadFromUDP(rda); err == nil {
			panic(err)
		} else if n > 3 {
			j := int(rda[1]) + int(rda[2])<<8
			if j >= blockSize/mtu {
				conn.WriteToUDP(nil, raddr)
			}
		}
	}
	fmt.Println(len(cReco))

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
