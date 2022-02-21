package utls

// 基于UDP的传输安全层交换密钥

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/lysShub/fudp/constant"
)

var ErrTLSHandshake = errors.New("tls handshake failed")
var ErrTLSHandshakeTimeout = errors.New("tls handshake timeout")
var ErrTLSHandshakeAuthority = errors.New("tls certificate signed by unknown authority")

// Server 交换密钥
func Server(conn *net.UDPConn, tlsCfg *tls.Config) (key [16]byte, err error) {

	cfg := &tls.Config{ClientAuth: tls.VerifyClientCertIfGiven, Certificates: tlsCfg.Certificates}
	sconn := NewSconn(conn)
	defer sconn.Close() // 不会close UDP Conn
	tconn := tls.Server(sconn, cfg)
	defer tconn.SetDeadline(time.Time{})
	defer func() { err = rewriteErr(err) }() // handle error

	if err = tconn.SetDeadline(time.Now().Add(constant.RTT << 4)); err != nil {
		return
	}
	if err = tconn.Handshake(); err != nil {
		return
	}
	if err = tconn.SetDeadline(time.Time{}); err != nil {
		return
	}

	var buf = make([]byte, constant.SIZE)
	if err = tconn.SetReadDeadline(time.Now().Add(constant.RTT << 1)); err != nil {
		return
	}
	var n int
	for {
		if n, err = tconn.Read(buf); err != nil {
			return
		} else {
			if n == constant.SIZE {
				copy(key[:], buf[:n])
				_, err = tconn.Write(buf[:n])
				return
			}
		}
	}
}

// Client
// selfRootCa支持自签证书
func Client(conn *net.UDPConn, key [16]byte, server string, rootCas ...*x509.Certificate) (err error) {
	// 安全信道建立后, 发送key并直到收到同一个key

	cfg := &tls.Config{
		CipherSuites: []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256},
		RootCAs:      x509.NewCertPool(),
		ServerName:   server,
	}
	for _, v := range rootCas {
		cfg.RootCAs.AddCert(v)
	}

	sconn := NewSconn(conn)
	defer sconn.Close() // 不会关闭raw conn, 注意buffer中是否还存在剩余数据
	tconn := tls.Client(sconn, cfg)
	defer tconn.SetDeadline(time.Time{})
	defer func() { err = rewriteErr(err) }() // handle error

	if err = tconn.SetDeadline(time.Now().Add(constant.RTT << 4)); err != nil {
		return
	}
	if err = tconn.Handshake(); err != nil {
		fmt.Println("handshake", err.Error())
		return
	}
	if err = tconn.SetDeadline(time.Time{}); err != nil {
		return
	}

	var buf []byte = make([]byte, constant.SIZE)
	if _, err = tconn.Write(key[:]); err != nil {
		return
	}
	if err = tconn.SetReadDeadline(time.Now().Add(constant.RTT << 1)); err != nil {
		return
	}
	var n int
	for {
		if n, err = tconn.Read(buf); err != nil {
			return
		} else if n == 16 && bytes.Equal(buf, key[:]) {
			return nil
		}
	}
}

func rewriteErr(err error) error {
	if err != nil {
		if serr := err.Error(); strings.Contains(serr, "authority") {
			err = ErrTLSHandshakeAuthority
		} else if strings.Contains(serr, "timeout") {
			err = ErrTLSHandshakeTimeout
		} else {
			err = ErrTLSHandshake
		}
	}
	return err
}

type sconn struct {
	// 使UDP Conn成为流式
	// 注意Close时：
	// 		1. 不会关闭原始的udp conn
	// 		2. buf中是否还存在剩余数据(当前忽略了此问题)

	conn *net.UDPConn

	buf  *bytes.Buffer
	rLen int
	lock sync.Mutex
}

func NewSconn(conn *net.UDPConn) *sconn {
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
		if err = s.push(&s.rLen); err != nil {
			return 0, err
		}
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
	// direct Write to raw conn
	s.lock.Lock()
	defer s.lock.Unlock()

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
