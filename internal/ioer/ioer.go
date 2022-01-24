package ioer

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/lysShub/fudp/constant"
	"github.com/lysShub/fudp/log"
)

// UDP路由(根据raddr)
// 当前只支持IPv4

const packetmtu uint16 = constant.MTU + 13 + 2 //
const readbuffer int = 4 << 10                 // 默认缓冲 4KB
type Listener struct {
	laddr *net.UDPAddr
	conn  *net.UDPConn

	buff [packetmtu]byte // 前两个字节存放数据长度
	m    map[int64]*Conn // 先使用map, 以后改为树
	sync.Mutex

	done              bool
	connDefaultBuffer int
}

type Conn struct {
	raddr *net.UDPAddr
	l     *Listener

	deadline *time.Ticker
	done     bool // 是否关闭

	buff [packetmtu]byte // 前2个字节是payload的长度
	spin spin

	// connected
	connected bool // 不会被改变
	conn      *net.UDPConn
	laddr     *net.UDPAddr
}

func Listen(network string, laddr *net.UDPAddr) (*Listener, error) {
	var l = new(Listener)
	var err error
	if l.conn, err = net.ListenUDP(network, laddr); err != nil {
		return nil, err
	}
	l.m = make(map[int64]*Conn, 64)
	l.laddr = laddr
	return l, nil
}

func (l *Listener) Close() error {
	l.Lock()
	defer l.Unlock()
	l.m = nil
	l.done = true
	return l.conn.Close()
}

func (l *Listener) Accept() (*Conn, error) {
	if l.done {
		return nil, errors.New("listener closed")
	}
	for {
		if n, raddr, err := l.conn.ReadFromUDP(l.buff[:]); err != nil {
			log.Error(err)
			continue
		} else {
			raddr.IP = raddr.IP.To16()

			// raddr默认文件IPv6格式 性能瓶颈
			v, ok := l.m[int64(raddr.IP[15])<<40|int64(raddr.IP[14])<<32|int64(raddr.IP[13])<<24|int64(raddr.IP[12])<<16|int64(raddr.Port)<<8|int64(raddr.Port)]
			if ok {
				l.Lock()
				n := copy(v.buff[2:], l.buff[:n])
				v.buff[0], v.buff[1] = byte(n>>8), byte(n)
				l.Unlock()
				v.spin.done() // 旧数据将被丢弃
			} else {
				// 新连接
				var c = Conn{}
				c.spin = make(spin, 1)
				c.l = l
				c.raddr = raddr

				l.Lock()
				l.m[int64(raddr.IP[15])<<40|int64(raddr.IP[14])<<32|int64(raddr.IP[13])<<24|int64(raddr.IP[12])<<16|int64(raddr.Port)<<8|int64(raddr.Port)] = &c

				n = copy(c.buff[2:], l.buff[:n])
				c.buff[0], c.buff[1] = byte(n>>8), byte(n)
				l.Unlock()
				c.spin.done()

				return &c, nil
			}
		}
	}
}

func Dial(network string, laddr, raddr *net.UDPAddr) (*Conn, error) {
	if uconn, err := net.DialUDP(network, laddr, raddr); err != nil {
		return nil, err
	} else {
		var c = new(Conn)
		c.raddr = raddr
		c.connected = true
		c.conn = uconn
		c.laddr = laddr
		return c, nil
	}
}

func (c *Conn) Read(d []byte) (int, error) {
	if c.done {
		return 0, errors.New("can't read from closed Conn")
	}

	if c.connected {
		return c.conn.Read(d)
	} else {
		if c.deadline != nil {
			select {
			case <-c.deadline.C:
				c.deadline.Reset(0) // deadline已过期, 必须重新设置; 由于select是随机的, 所以可能还是能读取到数据
				return 0, errors.New("io timeout")
			case <-c.spin:
				return copy(d, c.buff[2:2+int(c.buff[0])<<8|int(c.buff[1])]), nil
			}
		} else {
			c.spin.wait()
			return copy(d, c.buff[2:2+int(c.buff[0])<<8|int(c.buff[1])]), nil
		}
	}
}

func (c *Conn) Write(d []byte) (int, error) {
	if c.connected {
		return c.conn.Write(d)
	} else {
		return c.l.conn.WriteToUDP(d, c.raddr)
	}
}

func (c *Conn) Close() error {
	var err error
	if c.connected {
		err = c.conn.Close()
	} else {
		c.l.Lock()
		delete(c.l.m, int64(c.raddr.IP[15])<<40|int64(c.raddr.IP[14])<<32|int64(c.raddr.IP[13])<<24|int64(c.raddr.IP[12])<<16|int64(c.raddr.Port)<<8|int64(c.raddr.Port))
		c.l.Unlock()
	}
	c.done = true
	return err
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.SetReadDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	if c.connected {
		return c.conn.SetReadDeadline(t)
	} else {
		if t.IsZero() {
			c.deadline = nil
		} else {
			if c.deadline == nil {
				c.deadline = time.NewTicker(time.Until(t))
			} else {
				c.deadline.Reset(time.Until(t))
			}
		}
	}
	return nil
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	if c.connected {
		return c.conn.SetWriteDeadline(t)
	}
	return nil
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.raddr
}

func (c *Conn) RemoteUDPAddr() net.UDPAddr {
	return *c.raddr
}

func (c *Conn) LocalAddr() net.Addr {
	if c.connected {
		return c.laddr
	} else {
		return c.l.laddr
	}
}

func (c *Conn) LocalUDPAddr() net.UDPAddr {
	if c.connected {
		return *c.laddr
	} else {
		return *c.l.laddr
	}
}
