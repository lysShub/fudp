package fudp

/*
  握手
  握手时不可靠, 任何错误都会导致握手失败, 包括超时;

*/

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/lysShub/fudp/constant"
	"github.com/lysShub/fudp/internal/utls"
	"github.com/lysShub/fudp/packet"
)

// handPing 主动握手
// selfRootCa:	CS模式时, 使用自签证书需要设置
func (f *fudp) handPing(selfRootCa ...*x509.Certificate) (stateCode uint16, err error) {
	defer func() {
		f.rawConn.SetDeadline(time.Time{})
		if e := recover(); e != nil {
			err = errors.New("unknown handshake panic")
		}
	}()
	defer func() { f.rawConn.SetDeadline(time.Time{}) }()
	var da []byte = make([]byte, constant.MTU)

	if err := f.genKeyAndgcm(); err != nil {
		return 0, err
	}
	if f.isClient && !f.isP2P {
		if _, err = f.rawConn.Write(packet.Pack(da[:0], 0, 0, 0, nil)); err != nil {
			return 0, err
		}

		// tls 交换密钥
		if err = utls.Client(f.rawConn, f.key, f.url.Hostname(), selfRootCa...); err != nil {
			return 0, err
		}
	} else if f.isClient && f.isP2P {
		n := copy(da[0:], f.key[:])
		if _, err = f.rawConn.Write(packet.Pack(da[0:n], 0, 0, 0, f.gcm)); err != nil {
			return 0, err
		}
	} else {
		return 0, errors.New("unknown work mode")
	}
	// 密钥交换完成, 且gcm被初始化

	// 读取握手包1
	var n int
	if err = f.rawConn.SetReadDeadline(time.Now().Add(constant.RTT * 2)); err != nil {
		return 0, err
	}
	for {
		if n, err = f.rawConn.Read(da); err != nil {
			return 0, err
		} else {
			if ok, _ := f.expHandPackage(1, da[:n:cap(da)]); ok {
				break
			}
		}
	}
	if err = f.rawConn.SetDeadline(time.Time{}); err != nil {
		return 0, err
	}

	// 发送数据包2 即请求url
	burl := []byte(f.url.String())
	n = copy(da[0:], burl)
	if _, err = f.rawConn.Write(packet.Pack(da[0:n:cap(da)], 2, 0, 0, f.gcm)); err != nil {
		return 0, err
	}

	// 接受握手包3, statusCode
	if err = f.rawConn.SetReadDeadline(time.Now().Add(constant.RTT * 2)); err != nil {
		return 0, err
	}
	for {
		if n, err = f.rawConn.Read(da); err != nil {
			return 0, err
		} else {
			if ok, _ := f.expHandPackage(3, da[:n:cap(da)]); ok {
				return uint16(da[0]) + uint16(da[1])<<8, nil
			}
		}
	}
}

// handPong 接受握手
func (f *fudp) handPong() (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New("unknown handshake panic")
		}
	}()

	if !f.isClient && f.isP2P {
		if err = f.pongP2PSwapKey(); err != nil {
			return err
		}
	} else if !f.isClient && !f.isP2P {
		if err = f.pongCSSwapKey(); err != nil {
			return err
		}
	} else {
		return errors.New("unknown work mode")
	}
	if err = f.genKeyAndgcm(); err != nil {
		return err
	}
	// 已交换完成密钥 并初始化gcm

	var da []byte = make([]byte, constant.MTU)

	// 回复握手包1
	if _, err = f.rawConn.Write(packet.Pack(da[0:0:constant.MTU], 1, 0, 0, f.gcm)); err != nil {
		fmt.Println(err.Error())
		return err
	}

	var statueCode int // 状态码

	// 读取握手包2
	if err = f.rawConn.SetReadDeadline(time.Now().Add(constant.RTT << 1)); err != nil {
		return err
	}
	for {
		if n, err := f.rawConn.Read(da); err != nil {
			return err
		} else if n > 0 {
			if ok, da := f.expHandPackage(2, da[:n:cap(da)]); ok {
				if f.url, err = url.Parse(string(da)); err != nil {
					statueCode = http.StatusBadRequest
				} else {
					// handle func
					var rPath string
					if f.handleFn != nil {
						rPath, statueCode = f.handleFn(f.url)
					} else {
						if fn := Handle(f.url.Path); fn == nil {
							rPath, statueCode = "", http.StatusNotFound // 路由不存在
						} else {
							rPath, statueCode = fn(f.url)
						}
					}
					if f.wpath, err = filepath.Abs(rPath); err != nil {
						rPath, statueCode = "", http.StatusInternalServerError
					}

				}
				break
			}
		}
	}
	if err = f.rawConn.SetReadDeadline(time.Time{}); err != nil {
		return err
	}

	// 回复握手包3
	da[0], da[1] = byte(statueCode), byte(statueCode>>8)
	if _, err = f.rawConn.Write(packet.Pack(da[:2:cap(da)], 3, 0, 0, f.gcm)); err != nil {
		return err
	}

	if statueCode/100 == 2 {
		return nil
	}
	return errors.New("Status Code " + strconv.Itoa(statueCode))

}

// genKeyAndgcm 如果密钥为空则生成密钥; 并且初始化gcm实例
func (f *fudp) genKeyAndgcm() error {
	if tk := [16]byte{}; tk == f.key {
		if _, err := rand.Read(f.key[:]); err != nil {
			return err
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

// expHandPackage 判断是否时期望的数据包, 返回-1表示不是期望的数据包
func (f *fudp) expHandPackage(packageIndex int64, da []byte) (bool, []byte) {
	da, bi, other, pt, err := packet.Parse(da, f.gcm)
	if (err == nil) && (bi == packageIndex && other == 0 && pt == 0) {
		return true, da
	}
	return false, da
}

// pongP2PSwapKey server P2P模式交换密钥
func (f *fudp) pongP2PSwapKey() (err error) {

	defer func() {
		if e := f.rawConn.SetDeadline(time.Time{}); e != nil {
			err = e
		}
	}()

	var da []byte = make([]byte, constant.MTU)
	if err = f.rawConn.SetReadDeadline(time.Now().Add(constant.RTT << 1)); err != nil {
		return err
	}
	for {
		// 读取握手包0
		if n, err := f.rawConn.Read(da); err != nil {
			return err
		} else if n > 0 {
			if ok, _ := f.expHandPackage(0, da[:n]); ok {
				break
			} else {
				time.Sleep(constant.RTT >> 2)
			}
		}
	}
	return nil
}

func (f *fudp) pongCSSwapKey() (err error) {
	defer func() {
		if e := f.rawConn.SetDeadline(time.Time{}); e != nil {
			err = e
		}
	}()

	var da []byte = make([]byte, constant.MTU)

	if err = f.rawConn.SetReadDeadline(time.Now().Add(constant.RTT << 2)); err != nil {
		return err
	}
	for {
		if n, err := f.rawConn.Read(da); err != nil {
			return err
		} else {
			if ok, _ := f.expHandPackage(0, da[:n]); ok {
				break
			}
		}
	}

	f.key, err = utls.Server(f.rawConn, f.tlsCfg)
	return
}
