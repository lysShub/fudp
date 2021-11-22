package fudp

// 握手

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"time"

	"github.com/lysShub/fudp/constant"
	"github.com/lysShub/fudp/internal/crypter/cert"
	"github.com/lysShub/fudp/internal/crypter/ecc"
	"github.com/lysShub/fudp/packet"
)

// 握手期间超时时间
var handshakeTimeout = time.Millisecond * 200

const maxCap = constant.MTU + 13 // read buffer max capacity

// 接受握手
// 握手成功或失败都会导致函数退出
// 	@verify: 对于握手包0中参数的校验, 如果返回false则校验失败; 注意处理没有参数的情况
//  @timeout: 等待接收握手包0的超时时间
// 	@err: 返回错误, nil表示成功
//
func (f *Fudp) HandReceive(verify func(pars *url.URL) bool, timeout ...time.Duration) error {
	var err error
	var n int

	if len(timeout) != 0 {
		if err = f.conn.SetReadDeadline(time.Now().Add(timeout[0])); err != nil {
			return err
		}
	}
	defer f.conn.SetReadDeadline(time.Time{})

	var buff []byte = make([]byte, constant.MTU, maxCap)
	if n, err = f.conn.Read(buff); err != nil {
		return errors.New("数据读取错误: " + err.Error())
	} else {
		if length, fi, bias, pt, err := packet.Parse(buff[0:n], nil); err != nil {
			return err
		} else if fi == 0 && bias == 0 && pt == 0 {
		HANDHSAKE0:
			// 收到握手包0

			var rcode uint8
			var act uint8
			if length == 0 {
				rcode = 10 // 格式错误
			} else {
				var url *url.URL
				if length > 1 {
					if url, err = url.Parse(string(buff[1:])); err != nil {
						rcode = 11 // 参数格式错误
					}
				}

				if verify != nil && !verify(url) {
					rcode = 12 // 参数校验失败
				}

				act = buff[0]
				if rcode == 0 {
					if act != f.acti {
						rcode = 13 // 无权限请求
					}
				}
			}

			// 回复握手包1
			buff[0] = rcode
			var n int
			if f.mode == 1 && rcode < 10 {
				if len(buff)-1 < len(f.cert) {
					return errors.New("invalid cert that length more than " + strconv.Itoa(len(buff)-1))
				}
				n = copy(buff[1:], f.cert) + 1
			} else { // PP模式或拒绝请求将不回复证书
				n = 1
			}
			if length, err = packet.Pack(buff[:n], 1, 0, 0, nil); err != nil {
				return err
			}
			if _, err = f.conn.Write(buff[:length]); err != nil {
				return err
			}

			if rcode > 9 {
				return errors.New("response code: " + strconv.Itoa(int(rcode))) // 请求被拒绝、继续等待握手请求
			}

			// 等待接收握手包2
			if err = f.conn.SetReadDeadline(time.Now().Add(handshakeTimeout)); err != nil {
				return err
			}
			if n, err = f.conn.Read(buff); err != nil {
				return err
			} else {

				if length, fi, bias, pt, err := packet.Parse(buff[0:n], nil); err != nil {
					return err
				} else if length != 0 && fi == 0 && bias == 0 && pt == 0 {
					goto HANDHSAKE0 // 回退到握手包0
				} else if fi == 2 && bias == 0 && pt == 0 {

					// 收到握手包2
					if length < 16 {
						return errors.New("invalid length of handshake package index 2: " + strconv.Itoa(int(length))) // 握手包2格式错误
					} else {
						var pri []byte // 私钥解密
						if pri, err = cert.GetKeyInfo(f.key); err != nil {
							return err
						}

						if pt, err := ecc.Decrypt(pri, buff[:length]); err != nil {
							return err
						} else if n = copy(f.secretKey[:], pt); n != 16 || len(pt) != 16 {
							return errors.New("非对称解密错误：长度不正确： " + strconv.Itoa(n) + "  " + strconv.Itoa(len(pt)))
						}

						if block, err := aes.NewCipher(f.secretKey[:]); err != nil {
							return err
						} else {
							if f.gcm, err = cipher.NewGCM(block); err != nil {
								return err
							}
						}

						return nil // 握手成功

					}

				} else {
					// 非法的数据包
					return errors.New("invalid parameters of handshake package index 2: fi: " + strconv.Itoa(int(fi)) + " bias: " + strconv.Itoa(int(bias)) + " pt: " + strconv.Itoa(int(pt)))
				}
			}

		} else {
			// 非法的数据包
			return errors.New("invalid parameters of handshake package index 0: fi: " + strconv.Itoa(int(fi)) + " bias: " + strconv.Itoa(int(bias)) + " pt: " + strconv.Itoa(int(pt)))
		}
	}
}

// 发送握手
// 	@parmeters: 请求参数
// 	@err: 返回错误, nil表示成功
func (f *Fudp) HandSend(parmeters string) error {

	code := f.acti & 0b11
	if code == 0 || code == 3 {
		return errors.New("auth error")
	}

	var buff []byte = make([]byte, 0, maxCap)
	buff = append(buff[0:0:maxCap], code)
	buff = append(buff, parmeters...)

	defer f.conn.SetReadDeadline(time.Time{})

	if length, err := packet.Pack(buff, 0, 0, 0, nil); err != nil {
		return err
	} else {
		// 发送握手包0
		if _, err = f.conn.Write(buff[:length]); err != nil {
			return err
		} else {

			// 接收握手包1
			if f.conn.SetReadDeadline(time.Now().Add(handshakeTimeout)); err != nil {
				return err
			}
			buff = make([]byte, constant.MTU, maxCap)
			if n, err := f.conn.Read(buff[0:constant.MTU:maxCap]); err != nil {
				return err
			} else {
				if length, fi, bias, pt, err := packet.Parse(buff[:n], nil); err != nil {
					return errors.New("handshake package index 1: " + err.Error())
				} else if fi == 1 && bias == 0 && pt == 0 {
					if length < 1 {
						return errors.New("invalid length of handshake package index 1: " + strconv.Itoa(int(length))) // 握手包1格式错误
					}

					code := buff[0]
					if code > 9 { // 请求被拒绝
						return errors.New("response code:" + strconv.Itoa(int(code)))
					}
					ca := buff[1:]

					var publicKey []byte
					// 证书校验
					if f.mode == 0 { // PP
						publicKey = f.token

					} else { // CS client
						ok := false
						for _, v := range f.selfCert { // 自签
							if cert.VerifyCertificate(ca, v) == nil {
								ok = true
								break
							}
						}
						if !ok { // CA
							if err := cert.VerifyCertificate(ca); err != nil {
								return err
							}
						}

						if publicKey, _, _, err = cert.GetCertInfo(ca); err != nil {
							return errors.New("certificate parse fail: " + err.Error())
						}
					}

					// 回复握手包2
					var secretKey []byte = make([]byte, 16)
					rand.Read(secretKey)
					copy(f.secretKey[:], secretKey)
					if block, err := aes.NewCipher(secretKey); err != nil {
						return err
					} else {
						if f.gcm, err = cipher.NewGCM(block); err != nil {
							return err
						}
					}

					if ct, err := ecc.Encrypt(publicKey, f.secretKey[0:]); err != nil { // 公钥加密
						return err
					} else {
						if len(ct) > len(buff) {
							return errors.New("invalid secret that length more than " + strconv.Itoa(len(buff)))
						}
						n := copy(buff[0:len(ct):maxCap], ct)
						if length, err = packet.Pack(buff[:n:maxCap], 2, 0, 0, nil); err != nil {
							return err
						}

						fmt.Println("密钥", f.secretKey)

						// 回复握手包2
						if _, err = f.conn.Write(buff[:length]); err != nil {
							return err
						}

						return nil // 握手成功
					}
				} else {
					return errors.New("invalid parameters of handshake package index 1: fi: " + strconv.Itoa(int(fi)) + " bias: " + strconv.Itoa(int(bias)) + " pt: " + strconv.Itoa(int(pt)))
				}
			}

		}

	}

	return nil
}
