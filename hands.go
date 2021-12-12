package fudp

// 握手
// HandXxxx函数只有三种种情况会退出：握手成功和超时, 以及影响继续通信的错误、如Read/Write错误; 而诸如Package Parse错误后会restart。
// 因此：constant.HandshakeTimeout不要设置过大, 建议5RTT

// 握手是无状态的: C端局部超时会尝试回退重发。S端直接按照顺序执行

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lysShub/fudp/constant"
	"github.com/lysShub/fudp/internal/crypter/cert"
	"github.com/lysShub/fudp/internal/crypter/ecc"
	"github.com/lysShub/fudp/log"
	"github.com/lysShub/fudp/packet"
)

const maxCap = constant.MTU + 13 // read buffer max capacity

// HandPing 主动握手 C端
func (f *Fudp) HandPing() (stateCode uint16, err error) {
	var subReadTimeout = constant.HandshakeTimeout >> 2

	var urlPars string // url的参数

	if _, err := url.Parse(urlPars); log.Error(err) != nil {
		return 0, errors.New("bad url, unable to parse")
	}
	defer f.conn.SetReadDeadline(time.Time{})

	var start = time.Now()
	var buf []byte = make([]byte, maxCap)
	var n int
handsharkStart:
	if length, err := packet.Pack(buf[0:0:maxCap], 0, 0, 0, nil); log.Error(err) != nil {
		return 0, err
	} else {
		if _, err = f.conn.Write(buf[:length]); log.Error(err) != nil {
			return 0, ErrUnknown
		}
	}
	if err = f.conn.SetReadDeadline(time.Now().Add(subReadTimeout)); log.Error(err) != nil {
		return 0, ErrUnknown
	}
	if n, err = f.conn.Read(buf); log.Error(err) != nil {
		if time.Since(start) > constant.HandshakeTimeout {
			if strings.Contains(err.Error(), "timeout") {
				return http.StatusRequestTimeout, errors.New("request timeout")
			} else {
				return 0, ErrUnknown
			}
		} else {
			goto handsharkStart
		}
	} else {
		goto handsharkPong
	}

handsharkPong:
	if length, fi, bias, pt, err := packet.Parse(buf[:n], nil); log.Error(err) != nil || fi != 1 || bias != 0 || pt != 0 {
		goto handsharkStart
	} else {
		var publicKey *ecdsa.PublicKey
		if f.mode == CSMode {
			// CS 验签证书
			if err = cert.VerifyCertificate(buf[:length], f.cert); log.Error(err) != nil {
				return 495, errors.New("certificate verify failed")
			}
			if publicKey, err = cert.GetCertPubkey(buf[:length]); log.Error(err) != nil {
				if err == cert.ErrInvalidCertificateType {
					return 495, cert.ErrInvalidCertificateType
				}
				return 0, errors.New("get public key from certificate failed")
			}
		} else if f.mode == PPMode {
			if length != 0 {
				return http.StatusMethodNotAllowed, errors.New("work mode not match")
			}
			// PP 直接使用token 作为公钥
			if publicKey, err = ecc.ParsePubKey(f.token); log.Error(err) != nil {
				return 0, errors.New("invalid token")
			}
		} else {
			return 0, errors.New("unknown work mode")
		}

		var secretKey []byte = make([]byte, 16)
		rand.Read(secretKey)
		copy(f.secretKey[:], secretKey)
		if block, err := aes.NewCipher(secretKey); log.Error(err) != nil {
			return 0, ErrUnknown
		} else {
			if f.gcm, err = cipher.NewGCM(block); log.Error(err) != nil {
				return 0, ErrUnknown
			}
		}

		if ck, err := ecc.Encrypt(publicKey, f.secretKey[0:]); err != nil { // 公钥加密
			return 0, err
		} else {
			var cp []byte
			if len(urlPars) != 0 {
				cp = f.gcm.Seal(nil, make([]byte, 12), []byte(urlPars), nil)
			}
			m := copy(buf[0:], ck)
			m = copy(buf[m:], cp)
			if length, err := packet.Pack(buf[0:m+len(ck):maxCap], 1, 0, 0, nil); log.Error(err) != nil {
				return 0, err
			} else {
				if _, err = f.conn.Write(buf[:length]); err != nil {
					return 0, ErrUnknown
				} else {
					goto handsharkNext // 进入下一步
				}
			}
		}
	}

handsharkNext:
	if err = f.conn.SetReadDeadline(time.Now().Add(subReadTimeout)); log.Error(err) != nil {
		return 0, ErrUnknown
	}
	if n, err = f.conn.Read(buf); log.Error(err) != nil {
		if time.Since(start) > constant.HandshakeTimeout {
			if strings.Contains(err.Error(), "timeout") {
				return http.StatusRequestTimeout, errors.New("request timeout")
			} else {
				return 0, ErrUnknown
			}
		} else {
			goto handsharkStart
		}
	} else {
		if length, fi, bias, pt, err := packet.Parse(buf[:n], nil); log.Error(err) != nil {
			return 0, err
		} else {
			if length < 2 || fi != 0 || bias != 0 || pt != 3 {
				if fi != 1 || bias != 0 || pt != 0 {
					goto handsharkPong
				} else {
					goto handsharkStart
				}
			}
			fmt.Println(string(buf[2:length]))

			return uint16(buf[1])<<8 | uint16(buf[0]), nil
		}
	}
}

// 接受握手 S端
//  @waitTimeout: 等待握手开始超时时间, 默认4秒
// 	@err: 返回错误, nil表示握手成功
func (f *Fudp) HandPong(waitTimeout ...time.Duration) (stateCode uint16, err error) {

	defer f.conn.SetReadDeadline(time.Time{})
	if len(waitTimeout) != 0 && waitTimeout[0] > 0 {
		if err = f.conn.SetReadDeadline(time.Now().Add(waitTimeout[0])); log.Error(err) != nil {
			return 0, err
		}
	} else {
		if err = f.conn.SetReadDeadline(time.Now().Add(time.Second * 4)); log.Error(err) != nil {
			return 500, err
		}
	}
	var buf []byte = make([]byte, maxCap)
	var n int

	var start time.Time // 握手开始
waitStart:
	if n, err = f.conn.Read(buf); log.Error(err) != nil {
		if strings.Contains(err.Error(), "timeout") {
			return http.StatusRequestTimeout, errors.New("request timeout")
		} else {
			return 0, ErrUnknown
		}
	} else {
		if length, fi, bias, pt, err := packet.Parse(buf[0:n], nil); log.Error(err) != nil || length != 0 || fi != 0 || bias != 0 || pt != 0 {
			goto waitStart
		}
	}

	var got bool = false
handsharkStart:
	if start.IsZero() {
		start = time.Now()
	} else if !got {
		if err = f.conn.SetReadDeadline(time.Now().Add(constant.HandshakeTimeout - time.Since(start))); log.Error(err) != nil {
			return 0, ErrUnknown
		}
		if n, err = f.conn.Read(buf); log.Error(err) != nil {
			if time.Since(start) > constant.HandshakeTimeout {
				return http.StatusRequestTimeout, errors.New("request timeout")
			} else {
				return 0, ErrUnknown
			}
		}
	}

	if length, fi, bias, pt, err := packet.Parse(buf[0:n], nil); log.Error(err) != nil {
		goto handsharkStart
	} else if fi != 0 || bias != 0 || pt != 0 || length != 0 {
		goto handsharkStart
	}

handshakeNext:
	n = copy(buf[0:], f.cert)
	if length, err := packet.Pack(buf[0:n], 0, 0, 1, nil); log.Error(err) != nil {
		return 0, err
	} else {
		if _, err = f.conn.Write(buf[:length]); err != nil {
			return 0, ErrUnknown
		}
	}

	if err = f.conn.SetReadDeadline(time.Now().Add(constant.HandshakeTimeout - time.Since(start))); log.Error(err) != nil {
		return 0, ErrUnknown
	}
	if n, err = f.conn.Read(buf); log.Error(err) != nil {
		if time.Since(start) >= constant.HandshakeTimeout {
			return http.StatusRequestTimeout, errors.New("request timeout")
		} else {
			return 0, ErrUnknown
		}
	} else {

		if length, fi, bias, pt, err := packet.Parse(buf[0:n], nil); log.Error(err) != nil {
			goto handshakeNext
		} else if fi == 0 && bias == 0 && pt == 0 && length == 0 {
			got = true
			goto handsharkStart
		} else if fi == 0 && bias == 0 && pt == 2 {

			var rcode uint16 = 0
			var rmsg string = ""
			lk := uint16(buf[1])<<8 | uint16(buf[0])
			if lk+2 > length {
				rcode, rmsg = 400, "bad request, wrong format" // 包格式错误
			} else {
				var ck, cu []byte = make([]byte, lk), make([]byte, length-lk-2)
				copy(ck, buf[2:lk+2])
				copy(cu, buf[lk+2:length])

				if key, err := ecc.Decrypt(f.key, ck); err == nil && len(key) == 16 {
					copy(f.secretKey[:], key)

					if len(cu) > 0 {
						if urlPars, err := f.gcm.Open(nil, make([]byte, 12), cu, nil); err == nil {

							if url, err := url.Parse(string(urlPars)); err == nil {
								if err := f.verifyFunc(url); err != nil {
									rcode, rmsg = 401, err.Error() // 服务器拒绝 校验失败
								}
							} else {
								rcode, rmsg = 400, "bad request, unable to parse" // 请求信息错误 参数解密失败
							}
						} else {
							rcode, rmsg = 400, "bad request, unable to decrypt" // 请求信息错误 参数解密失败
						}
					}

					if rcode == 0 {
						if block, err := aes.NewCipher(f.secretKey[:]); err != nil {
							rcode, rmsg = 500, "fudp server error" // 服务器错误
						} else {
							if f.gcm, err = cipher.NewGCM(block); err != nil {
								rcode, rmsg = 500, "fudp server error" // 服务器错误
							}
						}
						rcode, rmsg = 400, ""
					}
				} else {
					rcode, rmsg = 400, "bad request, unable to parse" // 请求信息错误  对称加密密钥不合法
				}

				buf[0], buf[1] = byte(rcode), byte(rcode>>8)
				n := copy(buf[2:], []byte(rmsg))
				if n, err := packet.Pack(buf[0:n+2:maxCap], 3, 0, 0, nil); err == nil {
					if _, err = f.conn.Write(buf[:n]); err != nil {
						return 500, err
					}
				}

				// UDP不可靠，所以握手包2也可能不能正确被C端接收到，最后C端重试握手均失败后下线, S通过传输Timeout下线
				if rcode%100 == 4 {
					return rcode, nil
				} else {
					return rcode, errors.New(rmsg)
				}
			}

		} else {

		}

	}

	return 500, errors.New("") //
}

var ErrUnknown error = errors.New("unknown error")
