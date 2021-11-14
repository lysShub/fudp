package ioer

import (
	"errors"
	"net"
	"sync"
)

// UDP路由(根据raddr)

const MTU uint16 = 65535 //

type Ioer struct {
	laddr *net.UDPAddr
	conn  *net.UDPConn

	buff [MTU]byte

	// 先使用map, 以后改为树
	m map[int64]*Conn

	sync.Mutex
}

func New(laddr *net.UDPAddr) (*Ioer, error) {
	var i = new(Ioer)
	var err error
	if i.conn, err = net.ListenUDP("udp", laddr); err != nil {
		return nil, err
	}

	return nil, err
}

// Accept
func (i *Ioer) Accept() *Conn {

	for {
		if n, raddr, err := i.conn.ReadFromUDP(i.buff[:]); err != nil {
			// log
			continue
		} else {
			// raddr默认文件IPv6格式
			if v, ok := i.m[int64(raddr.IP[15])<<40|int64(raddr.IP[14])<<32|int64(raddr.IP[13])<<24|int64(raddr.IP[12])<<16|int64(raddr.Port)<<8|int64(raddr.Port)]; ok {
				if len(v.buff) == 0 { // BUG!!!!!!!!!!!!
					v.buffLock <- uint16(copy(v.buff[:], i.buff[:n]))
				}
			} else { //new
				i.Lock()
				var c = new(Conn)
				c.buffLock = make(chan uint16, 1)
				c.i = i
				c.raddr = raddr
				c.buffLock <- uint16(copy(c.buff[:], i.buff[:n]))
				i.Unlock()
				return c
			}
		}
	}
}

type Conn struct {
	buffLock chan uint16 // 其实只需要一个自旋锁
	buff     [MTU]byte
	i        *Ioer
	raddr    *net.UDPAddr
	done     bool // 是否关闭
}

func (c *Conn) Read(d []byte) (int, error) {
	if c.done {
		return 0, errors.New("can't read from closed Conn")
	}
	return copy(d, c.buff[:]), nil
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

func (c *Conn) Which() (laddr *net.UDPAddr, raddr *net.UDPAddr) {
	return c.i.laddr, c.raddr
}
