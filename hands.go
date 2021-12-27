package fudp

// fudp 握手

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/lysShub/fudp/constant"
	"github.com/lysShub/fudp/internal/crypter/cert"
	"github.com/lysShub/fudp/internal/crypter/ecc"
	"github.com/lysShub/fudp/log"
	"github.com/lysShub/fudp/packet"
)

const maxCap = constant.MTU + 13 // read buffer max capacity

var errHandshake error = errors.New("handshake failed")

// HandPing 主动握手 C端
// 	@stateCode: 状态码, 基本符合HTTPCode, 为0表示本地错误
// 	@err: 错误信息或S端reply msg
// 握手时没有可靠性保证, 任何数据错误将会导致握手失败、函数退出
func (f *fudp) HandPing(ctx context.Context, config Config) (stateCode uint16, err error) {
	defer f.conn.SetReadDeadline(time.Time{})
	if ctx != nil {
		var endCh = make(chan struct{})
		defer func() { endCh <- struct{}{} }()
		go func() {
			select {
			case <-ctx.Done():
			case <-endCh:
			}
			f.conn.SetReadDeadline(time.Now())
		}()

	}

	var buf []byte = make([]byte, maxCap)
	var n int

	if length, err := packet.Pack(buf[0:0:maxCap], 0, 0, 0, nil); log.Error(err) != nil {
		return 0, err
	} else {
		if _, err = f.conn.Write(buf[:length]); log.Error(err) != nil {
			return 0, err
		}
	}

	if err = f.conn.SetReadDeadline(time.Now().Add(constant.RTT + constant.RTT/2)); log.Error(err) != nil {
		return 0, errHandshake
	}
	if n, err = f.conn.Read(buf); log.Error(err) != nil {
		if strings.Contains(err.Error(), "timeout") {
			return http.StatusRequestTimeout, errors.New("timeout")
		} else {
			return 0, errHandshake
		}
	} else {
		if length, fi, bias, pt, err := packet.Parse(buf[0:n], nil); log.Error(err) != nil {
			return 0, errHandshake
		} else {
			if fi == 0 && bias == 0 && pt == 1 && length != 0 {
				var publicKey *ecdsa.PublicKey
				if config.mode == CSMode { // CS 验签证书
					if err = cert.VerifyCertificate(buf[:length], config.cert); log.Error(err) != nil {
						return 0, errHandshake
					}
					if publicKey, err = cert.GetCertPubkey(buf[:length]); log.Error(err) != nil {
						if err == cert.ErrInvalidCertificateType {
							return 0, errHandshake
						}
						return 0, errHandshake
					}
				} else if config.mode == PPMode { // PP 直接使用token 作为公钥
					if length != 0 {
						log.Error(errors.New("PP mode got data, length = " + strconv.Itoa(int(length))))
						return http.StatusMethodNotAllowed, errHandshake
					}
					if publicKey, err = ecc.ParsePubKey(config.token); log.Error(err) != nil {
						return 0, errHandshake
					}
				} else {
					return 0, errors.New("unknown work mode")
				}

				var secretKey []byte = make([]byte, 16)
				rand.Read(secretKey)
				copy(f.secretKey[:], secretKey)
				if block, err := aes.NewCipher(secretKey); log.Error(err) != nil {
					return 0, errHandshake
				} else {
					if f.gcm, err = cipher.NewGCM(block); log.Error(err) != nil {
						return 0, errHandshake
					}
				}

				if ck, err := ecc.Encrypt(publicKey, f.secretKey[0:]); log.Error(err) != nil { // 公钥加密
					return 0, errHandshake
				} else {
					u, err := url.Parse(config.url)
					if log.Error(err) != nil {
						return 0, errors.New("invalid url")
					}
					var cp []byte
					if len(u.RawQuery) != 0 {
						cp = f.gcm.Seal(nil, make([]byte, 12), []byte(u.RawQuery), nil)
					}
					m := copy(buf[0:], ck)
					m = copy(buf[m:], cp)
					if m+len(ck) >= constant.MTU {
						return 0, errors.New("url parameters too long")
					}
					if length, err := packet.Pack(buf[0:m+len(ck):maxCap], 1, 0, 0, nil); log.Error(err) != nil {
						return 0, err
					} else {
						if _, err = f.conn.Write(buf[:length]); err != nil {
							return 0, errHandshake
						}
						// 进入下一步
					}
				}

			} else {
				log.Error(errors.New("unexpected packet: fileIndex=" + strconv.Itoa(int(fi)) + ", bias=" + strconv.Itoa(int(bias)) + ", packageType=" + strconv.Itoa(int(pt)) + ", length=" + strconv.Itoa(int(length))))
				return 0, errHandshake
			}
		}
	}

	if err = f.conn.SetReadDeadline(time.Now().Add(constant.RTT + constant.RTT/2)); log.Error(err) != nil {
		return 0, errHandshake
	}
	if n, err = f.conn.Read(buf); log.Error(err) != nil {
		if strings.Contains(err.Error(), "timeout") {
			return http.StatusRequestTimeout, errors.New("timeout")
		} else {
			return 0, errHandshake
		}
	} else {
		if length, fi, bias, pt, err := packet.Parse(buf[0:n], nil); log.Error(err) != nil {
			return 0, errHandshake
		} else {
			if fi == 0 && bias == 0 && pt == 3 && length >= 2 {
				return uint16(buf[1])<<8 | uint16(buf[0]), errors.New(string(buf[2:length]))
			} else {
				log.Error(errors.New("unexpected packet: fileIndex=" + strconv.Itoa(int(fi)) + ", bias=" + strconv.Itoa(int(bias)) + ", packageType=" + strconv.Itoa(int(pt)) + ", length=" + strconv.Itoa(int(length))))
				return 0, errHandshake
			}
		}
	}
}

// 接受握手 S端
//  @waitTimeout: 等待握手开始超时时间, 默认4秒
// 	@err: 返回错误, nil表示握手成功
// 如果在握手中遇到错误将会重新开始, 函数只有三种情况会退出：握手成功、等待超时、致命错误
func (f *fudp) HandPong(ctx context.Context, config Config) (err error) {
	defer f.conn.SetReadDeadline(time.Time{})
	var buf []byte = make([]byte, maxCap)
	var n int

	var run, start time.Time = time.Now(), time.Time{} // 开始时间 握手开始时间
	var wait time.Duration = 1<<63 - 1                 // 等待握手开始超时时间
	if ctx != nil {
		if tout, ok := ctx.Deadline(); ok {
			wait = time.Until(tout)
		} else {
			var endCh = make(chan struct{})
			defer func() { endCh <- struct{}{} }()
			go func() {
				select {
				case <-ctx.Done():
				case <-endCh:
				}
				wait = time.Since(start)
				f.conn.SetReadDeadline(time.Now())
			}()
		}
	}

	for {
		if !start.IsZero() {
			if err = f.conn.SetReadDeadline(time.Now().Add(constant.RTT + constant.RTT/2)); log.Error(err) != nil {
				return errHandshake
			}
		} else {
			if err = f.conn.SetReadDeadline(time.Now().Add(wait - time.Since(run))); log.Error(err) != nil {
				return errHandshake
			}
		}

		if n, err = f.conn.Read(buf); log.Error(err) != nil {
			if !start.IsZero() {
				start = time.Time{}
				continue
			} else {
				if strings.Contains(err.Error(), "timeout") {
					return errors.New("timeout")
				} else {
					return errHandshake
				}
			}
		} else {
			if length, fi, bias, pt, err := packet.Parse(buf[0:n], nil); log.Error(err) != nil {
				continue
			} else if fi == 0 && bias == 0 && pt == 0 && length == 0 {
				n = copy(buf[0:], config.cert)
				if length, err := packet.Pack(buf[0:n], 0, 0, 1, nil); log.Error(err) != nil {
					return errHandshake
				} else {
					if _, err = f.conn.Write(buf[:length]); err != nil {
						return errHandshake
					} else {
						start = time.Now()
					}
				}

			} else if fi == 0 && bias == 0 && pt == 2 && length >= 2 && !start.IsZero() {
				var rcode uint16 = 0
				var rmsg string = ""
				lk := uint16(buf[1])<<8 | uint16(buf[0])
				if lk+2 > length {
					start = time.Time{}
					continue
				} else {
					var ck, cp []byte = make([]byte, lk), make([]byte, length-lk-2)
					copy(ck, buf[2:lk+2])
					copy(cp, buf[lk+2:length])

					if key, err := ecc.Decrypt(config.key, ck); log.Error(err) != nil {
						start = time.Time{}
						continue
					} else {
						if len(key) == 16 {
							log.Error(errors.New("unsupported key length: " + strconv.Itoa(len(key))))
							start = time.Time{}
							continue
						}
						var tmpGcm cipher.AEAD
						if block, err := aes.NewCipher(key); log.Error(err) != nil {
							return errHandshake
						} else {
							if tmpGcm, err = cipher.NewGCM(block); log.Error(err) != nil {
								return errHandshake
							}
						}

						var url *url.URL
						if len(cp) > 0 {
							if urlBytes, err := tmpGcm.Open(nil, make([]byte, 12), cp, nil); err == nil {
								if url, err = url.Parse(string(urlBytes)); err == nil {
									// if err := f.handleFunc(url); err != nil {
									// 	rcode, rmsg = 401, "url auth failed" // 服务器拒绝 校验失败
									// }
								} else {
									rcode, rmsg = 400, "url parse failed" // 请求信息错误 参数解密失败
								}
							} else {
								start = time.Time{}
								continue
							}
						}

						buf[0], buf[1] = byte(rcode), byte(rcode>>8)
						n := copy(buf[2:], []byte(rmsg))
						if n, err := packet.Pack(buf[0:n+2:maxCap], 3, 0, 0, nil); log.Error(err) != nil {
							return errHandshake
						} else {
							if _, err = f.conn.Write(buf[:n]); log.Error(err) != nil {
								return err
							}
						}

						copy(f.secretKey[:], key)
						f.gcm = tmpGcm
						config.url = url.String()
						return nil
					}

				}
			} else {
				if !start.IsZero() {
					start = time.Time{}
				}
				continue
			}
		}
	}
}
