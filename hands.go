package fudp

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/lysShub/fudp/constant"
	"github.com/lysShub/fudp/internal/sconn"
	"github.com/lysShub/fudp/packet"
)

// handshake
// 握手后得到conn
// 握手时不可靠, 任何错误都会导致握手失败, 包括超时; 且超时时间很短,只有2RTT

//
func (f *fudp) handPing() (stateCode uint16, err error) {
	mcap := constant.MTU + packet.Append
	var da []byte = make([]byte, 0, mcap)

	// ping
	if tk := [16]byte{}; f.key == tk {
		da = nil
	} else {
		var gcm cipher.AEAD
		if block, err := aes.NewCipher(f.key[:]); err != nil {
			return 0, err
		} else {
			if gcm, err = cipher.NewGCM(block); err != nil {
				return 0, err
			}
		}

		tda := gcm.Seal(nil, make([]byte, 12), f.key[:], nil)
		copy(tda[0:len(tda):cap(da)], tda)
	}

	if n, err := packet.Pack(da, 0, 0, 0, nil); err != nil {
		return 0, err
	} else {
		if _, err = f.rawConn.Write(da[:n]); err != nil {
			return 0, err
		}
	}

	// wait pong
	var n int
	if err = f.rawConn.SetReadDeadline(time.Now().Add(constant.RTT * 2)); err != nil {
		return 0, err
	}
	if n, err = f.rawConn.Read(da[0:]); err != nil {
		f.rawConn.SetReadDeadline(time.Time{})
		return 0, err
	} else {
		f.rawConn.SetReadDeadline(time.Time{})

		if len, fi, bi, pt, err := packet.Parse(da[:n], nil); err != nil {
			return 0, err
		} else if len != 1 || fi != 1 || bi != 0 || pt != 0 {
			return 0, errors.New("saafs")
		} else {
			if da[0] == 0 { // P2P mode
				//  nothing need do
			} else if da[0] == 1 { // CS mode
				rand.Read(f.key[:])
				if err = f.pingSwapSecert(); err != nil {
					return 0, err
				}
			} else {
				return 0, errors.New("handshake error, unexpect response")
			}
		}
	}
	if block, err := aes.NewCipher(f.key[:]); err != nil {
		return 0, err
	} else {
		if f.gcm, err = cipher.NewGCM(block); err != nil {
			return 0, err
		}
	}

	// request
	burl := []byte(f.url.String())
	copy(da[0:len(burl):cap(da)], burl)
	if len, err := packet.Pack(da, 2, 0, 0, f.gcm); err != nil {
		return 0, err
	} else {
		if _, err = f.rawConn.Write(da[:len]); err != nil {
			return 0, err
		}
	}

	// wait response
	if err = f.rawConn.SetReadDeadline(time.Now().Add(constant.RTT * 2)); err != nil {
		return 0, err
	}
	if n, err = f.rawConn.Read(da[0:mcap:mcap]); err != nil {
		f.rawConn.SetReadDeadline(time.Time{})
		return 0, err
	} else {
		f.rawConn.SetReadDeadline(time.Time{})

		if len, fi, bi, pt, err := packet.Parse(da[:n], nil); err != nil {
			return 0, err
		} else if len < 2 || fi != 3 || bi != 0 || pt != 0 {
			return 0, errors.New("saafs")
		} else {
			fmt.Println(string(da[2:]))
			return uint16(da[0]) + uint16(da[1])<<6, nil
		}
	}
}

// pingSwapSecert 交换密钥
func (f *fudp) pingSwapSecert() error {
	sconn := sconn.NewSconn(f.rawConn)
	tconn := tls.Client(sconn, &tls.Config{CipherSuites: []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256}})

	var buf []byte = make([]byte, 16)

	if _, err := tconn.Write(f.key[:]); err != nil {
		return err
	}
	if err := tconn.SetReadDeadline(time.Now().Add(constant.RTT)); err != nil {
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

func (f *fudp) handPong() error {

	if f.key != [16]byte{} && f.tlsConfig == nil {
		return f.handP2PPong()
	} else if f.key == [16]byte{} && f.tlsConfig != nil {
		return f.handCSPong()
	} else {
		return errors.New("invalid configure")
	}
}

func (f *fudp) handP2PPong() error {
	mlen := constant.MTU + packet.Append
	var da []byte = make([]byte, mlen)

	if err := f.rawConn.SetReadDeadline(time.Now().Add(constant.RTT)); err != nil {
		return err
	}
	if n, err := f.rawConn.Read(da); err != nil {
		f.rawConn.SetReadDeadline(time.Time{})
		return err
	} else if n <= 32 {
		f.rawConn.SetReadDeadline(time.Time{})
		return errors.New("sadfasfdsa")
	} else {
		f.rawConn.SetReadDeadline(time.Time{})

		if l, fi, bis, pt, err := packet.Parse(da[:n], f.gcm); err != nil {
			return err
		} else if fi != 0 || bis != 0 || pt != 0 {
			return errors.New("fsadsa")
		} else if !bytes.Equal(f.key[:], da[:l]) {
			n = copy(da[0:], []byte{})

			f.rawConn.Write(da[:n])
			return errors.New("dsafas")
		} else {
			n = copy(da[0:], []byte{})

			f.rawConn.Write(da[:n])
		}
	}

	if err := f.rawConn.SetReadDeadline(time.Now().Add(constant.RTT)); err != nil {
		return err
	}
	if n, err := f.rawConn.Read(da); err != nil {
		f.rawConn.SetReadDeadline(time.Time{})
		return err
	} else {
		f.gcm.Open(nil, make([]byte, 12), da[:n], nil)
	}

	return nil
}

func (f *fudp) handCSPong() error {
	return nil
}
