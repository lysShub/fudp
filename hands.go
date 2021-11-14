package fudp

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/lysShub/fudp/internal/crypter/cert"
	"github.com/lysShub/fudp/internal/crypter/ecc"
	"github.com/lysShub/fudp/packet"
)

// 握手期间超时时间, 发送/接受到握手包0时进入握手流程
var timeout = time.Millisecond * 200

// 接受握手 阻塞函数, 直到收到握手包
// 	@err: 返回错误, nil表示成功
// 	握手阻塞时, 如果需要终端, 可以在外部执行conn.Close()
func (f *fudp) HandReceive() error {

	var err error
	var n int
	for {
		var buff []byte = make([]byte, 65506, 65536)
		if n, err = f.conn.Read(buff); err != nil {
			return err
		} else {
			if length, fi, bias, pt, err := packet.Parse(buff[0:n], nil); err != nil {
				return err
			} else if length != 0 && fi == 0 && bias == 0 && pt == 0 {
			HANDHSAKE0:
				// 收到握手包0
				data := buff[:length]
				code := f.verifyAct(data)

				// 回复握手包1
				buff[0] = code
				buff = append(buff[1:1:65536], f.caCert...)
				if length, err = packet.Pack(buff[:len(f.caCert)+1], 1, 0, 0, nil); err != nil {
					return err
				}
				if _, err = f.conn.Write(buff[:length]); err != nil {
					return err
				}
				if code > 9 {
					continue // 请求被拒绝、继续等待握手请求
				}

				for {
					// 等待接收握手包2
					if err = f.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
						return err
					}
					if n, err = f.conn.Read(buff); err != nil {
						if err = f.conn.SetReadDeadline(time.Time{}); err != nil {
							return err
						}
						return err
					} else {
						if err = f.conn.SetReadDeadline(time.Time{}); err != nil {
							return err
						}

						if length, fi, bias, pt, err := packet.Parse(buff[0:n], nil); err != nil {
							return err
						} else if length != 0 && fi == 0 && bias == 0 && pt == 0 {
							goto HANDHSAKE0 // 回退到握手包0
						} else if length != 0 && fi == 2 && bias == 0 && pt == 0 {
							// 收到握手包2
							if pt, err := ecc.Decrypt(f.pubKey, buff[:length]); err != nil {
								return err
							} else if n = copy(f.secretKey[:], pt); n != 16 || len(pt) != 16 {
								return errors.New("非对称解密错误：长度不正确： " + strconv.Itoa(n) + "  " + strconv.Itoa(len(pt)))
							}
							return nil // 握手成功
						} else {
							// 非法的数据包
							fmt.Println("非法的数据包2")
							// return nil
							continue
						}
					}
				}

			} else {
				// 非法的数据包
				fmt.Println("非法的数据包")
				continue
			}
		}
	}
}

// 发送握手
// 	@err: 返回错误, nil表示成功
func (f *fudp) HandSend(actDonwload bool, data string) error {
	var code uint8
	if actDonwload {
		code = 1
	} else {
		code = 2
	}
	var buff []byte = make([]byte, 0, 65536)
	buff = append(buff, code)
	buff = append(buff, data...)

	if length, err := packet.Pack(buff, 0, 0, 0, nil); err != nil {
		return err
	} else {

		// 发送握手包0
		if _, err = f.conn.Write(buff[:length]); err != nil {
			return err
		} else {

			if n, err := f.conn.Read(buff[0:0:65536]); err != nil {
				return err
			} else {

				if f.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
					return err
				}
				if length, fi, bias, pt, err := packet.Parse(buff[:n], nil); err != nil {
					if f.conn.SetReadDeadline(time.Time{}); err != nil {
						return err
					}
					return err
				} else if length != 0 && fi == 1 && bias == 0 && pt == 0 {
					// 收到握手包1
					if f.conn.SetReadDeadline(time.Time{}); err != nil {
						return err
					}

					buff = buff[:length]
					code := buff[0]
					if code > 9 {
						// 请求被拒绝
						return nil
					}
					ca := buff[1:]
					if f.pubKey != nil { // 临时密钥
						pubkey, signature, data, err := cert.GetCertInfo(ca)
						if err != nil {
							return err
						}
						if !bytes.Equal(pubkey, f.pubKey) {
							return errors.New("非法的CA证书, 你可能找到网络攻击")
						}
						if ok, err := ecc.Verify(pubkey, signature, data); err != nil {
							return err
						} else if !ok {
							return errors.New("非法的CA证书, 你可能找到网络攻击")
						}

					} else if f.caCert != nil { // 自签
						if err := cert.VerifyCertificate(ca, f.caCert); err != nil {
							return err
						}
					} else { // CA验证
						if err := cert.VerifyCertificate(ca, nil); err != nil {
							return err
						}
					}

					// 回复握手包2
					var secretKey []byte = make([]byte, 16)
					rand.Read(secretKey)
					copy(buff[0:], secretKey)

					if ct, err := ecc.Encrypt(f.pubKey, f.secretKey[:]); err != nil {
						return err
					} else {

						n := copy(buff[0:0:65536], ct)
						if length, err = packet.Pack(buff[:n:65536], 2, 0, 0, nil); err != nil {
							return err
						}

						if _, err = f.conn.Write(buff[:length]); err != nil {
							return err
						}

						return nil // 握手成功
					}

				}

			}

		}

	}

	return nil
}
