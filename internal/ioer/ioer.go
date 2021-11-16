package ioer

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/lysShub/fudp"
)

// UDP路由(根据raddr)

const packetmtu uint16 = fudp.MTU + 13 + 2 //

type Ioer struct {
	laddr *net.UDPAddr
	conn  *net.UDPConn

	buff [packetmtu]byte // 前两个字节存放数据长度

	// 先使用map, 以后改为树
	m map[int64]*Conn

	sync.Mutex
}

func New(network string, laddr *net.UDPAddr) (*Ioer, error) {
	var i = new(Ioer)
	var err error
	if i.conn, err = net.ListenUDP(network, laddr); err != nil {
		return nil, err
	}
	i.m = make(map[int64]*Conn, 64)
	return i, nil
}

// Accept 阻塞函数、
//
func (i *Ioer) Accept() *Conn {

	for {
		if n, raddr, err := i.conn.ReadFromUDP(i.buff[:]); err != nil {
			// log
			continue
		} else {
			raddr.IP = raddr.IP.To16()
			// raddr默认文件IPv6格式
			v, ok := i.m[int64(raddr.IP[15])<<40|int64(raddr.IP[14])<<32|int64(raddr.IP[13])<<24|int64(raddr.IP[12])<<16|int64(raddr.Port)<<8|int64(raddr.Port)]
			if ok {
				// 可能出现在写v.buff的同时在读
				n := copy(v.buff[2:], i.buff[:n])
				v.buff[0], v.buff[1] = byte(n>>8), byte(n)
				v.spin.Signal()
				// 旧数据被丢弃

			} else { // 新连接
				i.Lock()
				var c = new(Conn)
				c.spin = &Spin{make(chan struct{}, 1)}
				c.i = i
				c.raddr = raddr
				i.m[int64(raddr.IP[15])<<40|int64(raddr.IP[14])<<32|int64(raddr.IP[13])<<24|int64(raddr.IP[12])<<16|int64(raddr.Port)<<8|int64(raddr.Port)] = c

				n = copy(c.buff[2:], i.buff[:n])
				c.buff[0], c.buff[1] = byte(n>>8), byte(n)

				i.Unlock()
				c.spin.Signal()
				return c
			}
		}
	}
}

type Conn struct {
	buff  [packetmtu]byte // 前2个字节是payload的长度
	i     *Ioer
	raddr *net.UDPAddr

	deadline *time.Ticker

	done bool // 是否关闭

	sync.Mutex
	spin *Spin
}

func (c *Conn) Read(d []byte) (int, error) {
	if c.done {
		return 0, errors.New("can't read from closed Conn")
	}

	if c.deadline != nil {
		select {
		case <-c.deadline.C:
			c.deadline.Reset(0) // deadline已过期, 必须重新设置; 由于select是随机的, 所以可能还是能读取到数据
			return 0, errors.New("io timeout")
		case <-*c.spin.WaitChan():
			return copy(d, c.buff[2:int(c.buff[0])<<8|int(c.buff[1])]), nil
		}
	} else {
		c.spin.Wait()
		return copy(d, c.buff[2:int(c.buff[0])<<8|int(c.buff[1])]), nil
	}
}

func (c *Conn) Write(d []byte) (int, error) {
	return c.i.conn.WriteToUDP(d, c.raddr)
}

func (c *Conn) Close() error {
	c.i.Lock()
	delete(c.i.m, int64(c.raddr.IP[15])<<40|int64(c.raddr.IP[14])<<32|int64(c.raddr.IP[13])<<24|int64(c.raddr.IP[12])<<16|int64(c.raddr.Port)<<8|int64(c.raddr.Port))
	c.done = true
	c.i.Unlock()
	return nil
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	if t.IsZero() {
		c.Lock()
		c.deadline = nil // t为time.Time{} 时取消deadline
		c.Unlock()
	} else {
		if c.deadline == nil {
			c.Lock()
			c.deadline = time.NewTicker(time.Until(t))
			c.Unlock()
		} else {
			c.deadline.Reset(time.Until(t))
		}
	}
	return nil
}

// SetWriteDeadline 无效函数！！！！！！！ 没有实现
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (c *Conn) RemoteAddr() net.UDPAddr {
	return *c.raddr
}

func (c *Conn) LocalAddr() net.UDPAddr {
	return *c.i.laddr
}
