package fudp

// 可靠文件(夹)传输协议

import (
	"net"

	"github.com/lysShub/fudp/packet"
)

type Fudp struct {
	// 表示一个Fudp通信

	Config

	conn net.Conn // connected UDP socket

	secretKey [16]byte   // 对称加密(AES_GCM_128)密钥
	gcm       packet.Gcm // 对称加密(AES_GCM_128)
}

type App struct {
	// 表示一个fudp应用程序, client或server

	// 不直到有没有用。。。。
}

// Run 启动
func Run(config Config, laddr, raddr string) (err error) {

	return nil
}

func Post(url string, path string) {}

func Put(url string, path string) {}
