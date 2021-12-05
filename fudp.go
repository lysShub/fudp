package fudp

// 可靠文件(夹)传输协议

import (
	"net"
	"net/url"

	"github.com/lysShub/fudp/packet"
)

// 定义
// Server端: 接受握手的一方
// Client端: 发送握手的一方

type Fudp struct {
	Config

	conn net.Conn // connected UDP socket

	secretKey [16]byte   // 对称加密(AES_GCM_128)密钥
	gcm       packet.Gcm // 对称加密(AES_GCM_128)
}

// Run 启动
func Run(config Config, conn net.Conn) (f *Fudp, err error) {

	return &Fudp{config, conn, [16]byte{}, nil}, nil
}

// Verify 请求参数校验, 校验通过返回0, 否则返回对应HTTP Code
func (f *Fudp) Verify(url *url.URL) (uint16, string) {

	// 校验资源是否存在, 以及请求中的保留参数等
	if false {
		return 404, "not found"
	}

	if !f.verifyFunc(url) {
		return 403, "forbidden" // forbidden
	}
	return 0, ""
}
