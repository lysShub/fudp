package fudp

// 握手
// UDP不可靠, 握手时所有数据包也不可靠。因此握手中的4个数据包中的任何一个发生丢包
// 或者bit错误，都会导致握手失败，而且只能通过Timeout结束；因此HandshakeTimeout不应设置为3~4RTT。
// 当握手失败时可以通过重试避免误判

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/lysShub/fudp/constant"
	"github.com/lysShub/fudp/internal/crypter/cert"
	"github.com/lysShub/fudp/internal/crypter/ecc"
	"github.com/lysShub/fudp/packet"
)

const maxCap = constant.MTU + 13 // read buffer max capacity

// 接受握手
//  @timeout: 等待握手超时时间, 默认4秒
// 	@err: 返回错误, nil表示握手成功
func (f *Fudp) HandReceive(timeout ...time.Duration) (stateCode uint16, err error) {

	defer f.conn.SetReadDeadline(time.Time{})

	if len(timeout) != 0 && timeout[0] > 0 {
		if err = f.conn.SetReadDeadline(time.Now().Add(timeout[0])); err != nil {
			return 500, err
		}
	} else {
		if err = f.conn.SetReadDeadline(time.Now().Add(constant.WaitHandShakeTimeout)); err != nil {
			return 500, err
		}
	}

	var buf []byte = make([]byte, maxCap)
	for i := 1; i <= constant.HandshakeRetry; i++ {

		var expStep uint16 = 0
		var start time.Time = time.Now()
		if n, err := f.conn.Read(buf); err == nil {
			if length, fi, bias, pt, err := packet.Parse(buf[0:n], &f.gcm); err == nil {

				if err = f.conn.SetReadDeadline(time.Now().Add(constant.HandshakeTimeout - time.Since(start))); err != nil {
					return 500, err
				}

				if expStep == 0 && fi == 0 && bias == 0 && pt == 0 && length == 0 {
					// 握手包0

					n = 0
					if f.mode == 1 {
						n = copy(buf[0:], f.cert)
					}
					if n, err := packet.Pack(buf[0:n:maxCap], 1, 0, 0, nil); err == nil {
						if _, err = f.conn.Write(buf[:n]); err != nil {
							return 500, err
						}
					} else {
						continue
					}
					expStep = 2
				} else if expStep == 2 && fi == 2 && bias == 0 && pt == 0 {
					// 握手包2

					var rcode uint16 = 0
					var rmsg string = ""
					lk := uint16(buf[1])<<8 | uint16(buf[0])
					if lk+2 > length {
						rcode, rmsg = 400, "bad request, wrong format" // 包格式错误
					} else {
						var ck, cu []byte = make([]byte, lk), make([]byte, length-lk-2)
						copy(ck, buf[2:lk+2])
						copy(cu, buf[lk+2:length])

						if len(f.key) == 0 {
							rcode, rmsg = 500, "fudp server error" // 获取私钥信息失败 服务器错误
						} else {
							if key, err := ecc.Decrypt(f.key, ck); err == nil && len(key) == 16 {
								copy(f.secretKey[:], key)

								if len(cu) > 0 {
									if purl, err := f.gcm.Open(nil, make([]byte, 12), cu, nil); err == nil {

										if url, err := url.Parse(string(purl)); err == nil {
											if r, msg := f.Verify(url); r != 0 {
												rcode, rmsg = r, msg // 服务器拒绝 校验失败
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
					continue
				}
			} else {
				continue
			}
		} else {
			if expStep == 0 && strings.Contains(err.Error(), "timeout") {
				return 500, errors.New("wait handshake timeout") // 等待握手超时
			} else if strings.Contains(err.Error(), "timeout") {
				continue
			}
			return 500, err
		}
	}
	return 500, errors.New("") //
}

func (f *Fudp) HandSend(purl string) (stateCode uint16, err error) {
	act := f.acti & 0b11
	if act == 0 || act == 3 {
		return 0, errors.New("auth error")
	}
	if _, err := url.Parse(purl); err != nil {
		return 0, errors.New("bad url, unable to parse")
	}
	defer f.conn.SetReadDeadline(time.Time{})

handsharkStart:
	var start time.Time = time.Now()
	var buf []byte = make([]byte, maxCap)
	buf[0] = act
	if length, err := packet.Pack(buf[0:1:maxCap], 0, 0, 0, nil); err != nil {
		return 0, err
	} else {
		if _, err = f.conn.Write(buf[:length]); err != nil {
			return 0, err
		}
	}
	for {
		if err = f.conn.SetReadDeadline(time.Now().Add(constant.HandshakeTimeout - time.Since(start))); err != nil {
			return 0, err
		}
		if n, err := f.conn.Read(buf); err == nil {

			if length, fi, bias, pt, err := packet.Parse(buf[:n], nil); err != nil || fi != 1 || bias != 0 || pt != 0 {
				goto handsharkStart // S端回复的数据不可能解包失败, 仅有可能发生了bit错误或遇到攻击
			} else {
				if f.mode == 1 && length == 0 {
					// 模式不匹配 C:CS  S:PP
					return 0, errors.New("mode not match")
				} else {
					var pubKey []byte
					if length != 0 {
						// 验签证书
						var ok bool = false
						for i := 0; i < len(f.selfCert) && !ok; i++ {
							if cert.VerifyCertificate(buf[:length], f.selfCert[i]) == nil {
								ok = true
							}
						}
						if !ok {
							if cert.VerifyCertificate(buf[:length]) == nil {
								ok = true
							}
						}

						if !ok {
							return 888, errors.New("certificate verify failed")
						}
						if pubKey, err = cert.GetCertPubkey(buf[:length]); err != nil {
							goto handsharkStart
						}
					} else {
						pubKey = f.token
					}

					var secretKey []byte = make([]byte, 16)
					rand.Read(secretKey)
					copy(f.secretKey[:], secretKey)
					if block, err := aes.NewCipher(secretKey); err != nil {
						return 500, err
					} else {
						if f.gcm, err = cipher.NewGCM(block); err != nil {
							return 500, err
						}
					}
					if ck, err := ecc.Encrypt(pubKey, f.secretKey[0:]); err != nil { // 公钥加密
						return 500, err
					} else {
						var cp []byte
						if len(purl) != 0 {
							cp = f.gcm.Seal(nil, make([]byte, 12), []byte(purl), nil)
						}
						n = copy(buf[0:], ck)
						n = copy(buf[n:], cp)
						if length, err := packet.Pack(buf[0:n+len(ck):maxCap], 1, 0, 0, nil); err != nil {

							if _, err = f.conn.Write(buf[:length]); err != nil {
								return 500, err
							} else {
								break
							}
						}
					}
				}
			}

		} else {
			return 500, err
		}
	}

	for {
		if err = f.conn.SetReadDeadline(time.Now().Add(constant.HandshakeTimeout - time.Since(start))); err != nil {
			return 500, err
		}
		if n, err := f.conn.Read(buf); err == nil {

			if length, fi, bias, pt, err := packet.Parse(buf[:n], nil); err != nil {
				continue
			} else {
				fmt.Println(length, fi, bias, pt)
				return 500, errors.New("")
			}

		}
	}
}
