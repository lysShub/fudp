package fudp

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"time"

	"github.com/lysShub/fudp/constant"
	"github.com/lysShub/fudp/internal/sconn"
	"github.com/lysShub/fudp/packet"
)

// handshake
// 握手后得到conn
// 握手时不可靠, 任何错误都会导致握手失败, 包括超时; 且超时时间很短,只有2RTT
// 握手发生失败时直接退出即可

const mcap = constant.MTU + packet.Append

//
func (f *fudp) handPing() (stateCode uint16, err error) {
	var da []byte = make([]byte, mcap)
	defer func() {
		if e := f.rawConn.SetReadDeadline(time.Time{}); e != nil {
			err = e
		}
	}()

	// ping
	if err := f.genSecretKey(); err != nil {
		return 0, err
	}
	if tk := [16]byte{}; f.key == tk {
		// CS
		if n, err := packet.Pack(da[:0:cap(da)], 0, 0, 0, nil); err != nil {
			return 0, err
		} else {
			if _, err = f.rawConn.Write(da[:n]); err != nil {
				return 0, err
			}
		}
		// // tls 交换密钥
		if _, err := rand.Read(f.key[:]); err != nil {
			return 0, err
		}
		if err := f.pingSwapSecertOverTLS(nil); err != nil {
			return 0, err
		}

	} else {
		// P2P
		n := copy(da[0:], f.key[:])
		if n, err := packet.Pack(da[0:n:cap(da)], 0, 0, 0, f.gcm); err != nil {
			return 0, err
		} else {
			if _, err = f.rawConn.Write(da[:n]); err != nil {
				return 0, err
			}
		}
	}

	// wait pong 读取数据包1
	var n int
	if err = f.rawConn.SetReadDeadline(time.Now().Add(constant.RTT * 2)); err != nil {
		return 0, err
	}
	for {
		if n, err = f.rawConn.Read(da); err != nil {
			return 0, err
		} else {
			if len, fi, bi, pt, err := packet.Parse(da[:n:cap(da)], nil); err != nil {
				return 0, err
			} else if len != 0 || fi != 1 || bi != 0 || pt != 0 {
				return 0, errors.New("saafs")
			} else {
				break
			}
		}
	}
	/*
	 * determined secret key
	 */

	// request 发送数据包2 及请求url
	burl := []byte(f.url.String())
	n = copy(da[0:], burl)
	if len, err := packet.Pack(da[0:n:cap(da)], 2, 0, 0, f.gcm); err != nil {
		return 0, err
	} else {
		if _, err = f.rawConn.Write(da[:len]); err != nil {
			return 0, err
		}
	}

	// wait response 接受握手包3, statusCode
	if err = f.rawConn.SetReadDeadline(time.Now().Add(constant.RTT * 2)); err != nil {
		return 0, err
	}
	if n, err = f.rawConn.Read(da); err != nil {
		return 0, err
	} else {
		if len, fi, bi, pt, err := packet.Parse(da[:n], f.gcm); err != nil {
			return 0, err
		} else if len != 2 || fi != 3 || bi != 0 || pt != 0 {
			return 0, errors.New("saafs")
		} else {
			fmt.Println(string(da[2:]))
			return uint16(da[0]) + uint16(da[1])<<8, nil
		}
	}
}

func (f *fudp) handPong() (err error) {

	if f.key != [16]byte{} && f.cfg == nil {
		if err = f.pongP2PSwapKey(); err != nil {
			return err
		}
	} else if f.key == [16]byte{} && f.cfg != nil {
		if err = f.pongCSSwapKey(); err != nil {
			return err
		}
	} else {
		return errors.New("invalid configure")
	}
	if tk := [16]byte{}; tk == f.key {
		return errors.New("未交换密钥")
	}
	// 已交换完成密钥

	var da []byte = make([]byte, mcap)
	//
	// 发送握手包1

	if n, err := packet.Pack(da[0:0:mcap], 1, 0, 0, nil); err != nil {
		return err
	} else {
		fmt.Println("握手包1", da[:n])
		if _, err = f.rawConn.Write(da[:n]); err != nil {
			return err
		}
	}

	// 读取握手包2
	if err := f.rawConn.SetReadDeadline(time.Now().Add(constant.RTT << 1)); err != nil {
		return err
	}
	for {
		if n, err := f.rawConn.Read(da); err != nil {
			return err
		} else if n > 0 {
			if len, fi, bi, pt, err := packet.Parse(da[:n], f.gcm); err != nil {
				return err
			} else if fi != 2 || bi != 0 || pt != 0 {
				time.Sleep(constant.RTT >> 2)
				continue
			} else {
				if f.url, err = url.Parse(string(da[:len])); err != nil {
					return err
				} else {
					break
				}
			}
		}
	}

	// handle
	var code int
	var rPath string
	if f.handlFn != nil {
		rPath, code = f.handlFn(f.url)
	} else {
		if fn := Handle(f.url.Path); fn == nil {
			rPath, code = "", http.StatusNotFound
		} else {
			rPath, code = fn(f.url)
		}
	}
	if !path.IsAbs(rPath) {
		rPath = path.Join(f.path, rPath) // f.path默认值为根路径
		if !path.IsAbs(rPath) {
			rPath, err = filepath.Abs(rPath)
			if err != nil {
				return err
			}
		}
	}
	f.path = rPath

	// 回复 握手包3
	da[0], da[1] = byte(code), byte(code>>8)
	if n, err := packet.Pack(da[:2:cap(da)], 3, 0, 0, f.gcm); err != nil {
		return err
	} else {
		if _, err = f.rawConn.Write(da[:n]); err != nil {
			return err
		}
	}

	if code/100 == 2 {
		return nil
	}
	return errors.New(" ")
}

func (f *fudp) pongP2PSwapKey() (err error) {
	if tk := [16]byte{}; f.key == tk {
		return errors.New("传输密钥不能为空")
	}
	if f.gcm == nil {
		if block, err := aes.NewCipher(f.key[:]); err != nil {
			return err
		} else {
			if f.gcm, err = cipher.NewGCM(block); err != nil {
				return err
			}
		}
	}
	defer func() {
		if e := f.rawConn.SetReadDeadline(time.Time{}); e != nil {
			err = e
		}
	}()

	var da []byte = make([]byte, mcap)
	if err = f.rawConn.SetReadDeadline(time.Now().Add(constant.RTT << 1)); err != nil {
		return err
	}
	for {
		// 读取握手包0
		if n, err := f.rawConn.Read(da); err != nil {
			return err
		} else if n > 0 {
			if len, fi, bi, pt, err := packet.Parse(da[:n], f.gcm); err != nil {
				return err
			} else {
				if len != constant.SIZE || fi != 0 || bi != 0 || pt != 0 {
					time.Sleep(constant.RTT >> 2)
					continue
				} else {
					break
				}
			}
		}
	}
	return nil
}

func (f *fudp) pongCSSwapKey() (err error) {
	defer func() {
		if e := f.rawConn.SetReadDeadline(time.Time{}); e != nil {
			err = e
		}
	}()
	var da []byte = make([]byte, mcap)

	if err = f.rawConn.SetReadDeadline(time.Now().Add(constant.RTT << 2)); err != nil {
		return err
	}
	for {
		if n, err := f.rawConn.Read(da); err != nil {
			return err
		} else {
			if len, fi, bi, pt, err := packet.Parse(da[:n], f.gcm); err != nil {
				return err
			} else {
				if len == 0 && fi == 0 && bi == 0 && pt == 0 {
					break
				}
			}
		}
	}

	sconn := sconn.NewSconn(f.rawConn)
	tconn := tls.Server(sconn, f.cfg)
	if err := tconn.SetDeadline(time.Now().Add(constant.RTT << 3)); err != nil {
		return err
	}
	tconn.Handshake()
	if err = f.pongSwapSecertOverTLS(); err != nil {
		return err
	}
	if err = tconn.Close(); err != nil {
		return err
	}
	return
}

// genSecretKey 如果密钥为空则生成密钥; 并且初始化gcm实例
func (f *fudp) genSecretKey() error {
	if tk := [16]byte{}; tk == f.key {
		if n, err := rand.Read(f.key[:]); err != nil {
			return err
		} else if n != constant.SIZE {
			return errors.New("生成密钥长度不正确")
		}
	}

	if block, err := aes.NewCipher(f.key[:]); err != nil {
		return err
	} else {
		if f.gcm, err = cipher.NewGCM(block); err != nil {
			return err
		}
	}
	return nil
}

// pingSwapSecertOverTLS 基于TLS交换密钥,用于CS模式Client
func (f *fudp) pingSwapSecertOverTLS(selfRootCa []*x509.Certificate) error {
	cfg := &tls.Config{
		CipherSuites: []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256},
		RootCAs:      x509.NewCertPool(),
	}
	for _, v := range selfRootCa {
		cfg.RootCAs.AddCert(v)
	}

	sconn := sconn.NewSconn(f.rawConn)
	tconn := tls.Client(sconn, cfg)
	if err := tconn.SetDeadline(time.Now().Add(constant.RTT << 3)); err != nil {
		return err
	}
	if err := tconn.Handshake(); err != nil {
		return err
	}

	var buf []byte = make([]byte, constant.SIZE)
	if _, err := tconn.Write(f.key[:]); err != nil {
		return err
	}
	if err := tconn.SetReadDeadline(time.Now().Add(constant.RTT << 1)); err != nil {
		return err
	}
	for {
		if n, err := tconn.Read(buf); err != nil {
			return err
		} else if n != 16 {
			continue
		} else {
			if bytes.Equal(buf, f.key[:]) {
				return nil
			} else {
				break
			}
		}
	}
	return errors.New("handshake timeout")
}

func (f *fudp) pongSwapSecertOverTLS() error {
	cfg := &tls.Config{ClientAuth: tls.VerifyClientCertIfGiven, Certificates: f.cfg.Certificates}
	sconn := sconn.NewSconn(f.rawConn)
	tconn := tls.Server(sconn, cfg)
	defer tconn.Close()

	if err := tconn.SetDeadline(time.Now().Add(constant.RTT << 8)); err != nil {
		return err
	}
	if err := tconn.Handshake(); err != nil {
		return err
	}

	var buf = make([]byte, constant.SIZE)
	if err := tconn.SetReadDeadline(time.Now().Add(constant.RTT << 1)); err != nil {
		return err
	}
	for {
		if n, err := tconn.Read(buf); err != nil {
			return err
		} else {
			if n == constant.SIZE {
				copy(f.key[:], buf[:n])
				tconn.Write(f.key[:])
				tconn.Write(f.key[:])
			} else {
				continue
			}
		}
	}
}
