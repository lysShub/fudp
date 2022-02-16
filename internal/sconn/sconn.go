package sconn

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// 将UDP Conn转换为流式Conn, 为了TLS基于UDP

type sconn struct {
	conn net.Conn

	buf  *bytes.Buffer
	rLen int
	lock sync.Mutex
}

func NewSconn(conn net.Conn) *sconn {
	return &sconn{
		conn: conn,

		buf:  bytes.NewBuffer(nil),
		rLen: 2000,
		lock: sync.Mutex{},
	}
}

func (s *sconn) Read(b []byte) (n int, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.buf.Len() != 0 {
		n, err = s.buf.Read(b)
	} else {
		s.push(&s.rLen)
		n, err = s.buf.Read(b)

	}
	return n, err
}

// push push UDP package to buff
func (s *sconn) push(rLen *int) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if *rLen < 65536 {
				*rLen = *rLen + 2000
				s.push(rLen)
			} else {
				err = errors.New(fmt.Sprintln(e))
			}
		}
	}()
	var tmp []byte = make([]byte, *rLen)
	if n, err := s.conn.Read(tmp); err != nil {
		return err
	} else {
		_, err := s.buf.Write(tmp[:n])
		return err
	}
}

func (s *sconn) Write(b []byte) (n int, err error) {

	for i := 0; i < len(b); i = i + 512 {
		j := i + 512
		if j > len(b) {
			j = len(b)
		}
		if n, err = s.conn.Write(b[i:j]); err != nil {
			return n, err
		}
	}
	return len(b), nil
}

func (s *sconn) Close() error                       { return nil } // only close TLS conn, don't close UDP conn
func (s *sconn) LocalAddr() net.Addr                { return s.conn.LocalAddr() }
func (s *sconn) RemoteAddr() net.Addr               { return s.conn.RemoteAddr() }
func (s *sconn) SetDeadline(t time.Time) error      { return s.conn.SetDeadline(t) }
func (s *sconn) SetReadDeadline(t time.Time) error  { return s.conn.SetReadDeadline(t) }
func (s *sconn) SetWriteDeadline(t time.Time) error { return s.conn.SetWriteDeadline(t) }
