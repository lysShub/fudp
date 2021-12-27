package fudp

// 可靠文件(夹)传输协议

import (
	"context"
	"errors"
	"net"
	"net/url"

	"github.com/lysShub/fudp/internal/crypter/cert"
	"github.com/lysShub/fudp/internal/crypter/ecc"
	"github.com/lysShub/fudp/internal/ioer"
	"github.com/lysShub/fudp/log"
	"github.com/lysShub/fudp/packet"
)

type fudp struct {
	// 一个fudp表示一个通信

	conn      net.Conn   // connected UDP socket
	secretKey [16]byte   // 对称加密(AES_GCM_128)密钥
	gcm       packet.Gcm // 对称加密(AES_GCM_128)
}

// ListenAndServer 必须是CS模式的Server端
func ListenAndServer(addr string, path string, ca, key []byte) (err error) {
	var conf Config

	if path, err = formatPath(path); err != nil {
		return err
	} else if err = checkSendPath(path, true); err != nil {
		return err
	}
	if len(ca) == 0 || !cert.CheckCertFormat(ca) {
		return errors.New("invalid CA certificate")
	}
	if _, err := ecc.ParsePriKey(key); len(key) == 0 || err != nil {
		return errors.New("invalid key")
	} else {
		var handle Handler = func(url *url.URL) (path string, err error) {
			return "", nil
		}

		conf, err = Configure(func(c *Config) {
			c.CSMode().Server(ca, key, handle)
		})
		if err != nil {
			return err
		}
	}

	var l *ioer.Listener
	if a, err := net.ResolveUDPAddr("udp", addr); err != nil {
		return err
	} else {
		if l, err = ioer.Listen("udp", a); err != nil {
			return err
		}
	}

	for {
		if conn, err := l.Accept(); err != nil {
			log.Error(err)
		} else {
			go func(conn net.Conn) {
				var f = fudp{conn: conn}
				if err := f.HandPong(context.Background(), conf); err != nil {
					return
				}
				// 开始传输

			}(conn)
		}
	}

	return
}

// FileServer 作为文件下载服务器启动服务
// 	@path 文件/文件夹路径, 必须存在
// 	@certPath 证书文件路径
// 	@keyPaht 密钥文件路径
func FileServer(addr, path, certPath, keyPath string) (err error) { return }

// Post 从服务器下载文件
// 	@url 请求地址
// 	@path 请求文件/文件夹在本地存放路径
// 	@ca 验签证书, 相较于系统证书优先使用
func Pull(ctx context.Context, url string, path string, ca ...[]byte) (err error) { return }

// Push 上传文件/文件夹到服务器
//	@url 请求地址
//	@path 上传的文件/文件夹路径
func Push(ctx context.Context, url string, path string) (err error) { return }

// Send 点对点模式发送文件
//	@path: 被发送文件/文件夹路径
//	@token: 本次传输的token, 接收方需要输入此token
func Send(ctx context.Context, path string, token *string) (err error) {
	// var p = &fudp{}
	// p.cert, p.key, err = cert.GenerateCert(time.Hour*24, func(c *cert.Csr) {})
	// if err != nil {
	// 	p.err = err
	// 	return
	// }

	// p.token, err = ecc.MarshalPubKey(&p.key.PublicKey)
	// if err != nil {
	// 	p.err = err
	// 	return
	// }
	return
}

// Receive 点对点模式接收文件
//	@path: 接收文件/文件夹在本地存放路径
// 	@token: 传输token
func Receive(ctx context.Context, url string, path string, token string) (err error) {
	// var p = &fudp{}
	// if p.key, p.err = ecc.GenerateKey(); p.err != nil {
	// 	return
	// }
	// if pk, err := ecc.MarshalPubKey(&p.key.PublicKey); err != nil {
	// 	p.err = err
	// 	return err
	// } else {
	// 	p.token = pk
	// }
	return
}
