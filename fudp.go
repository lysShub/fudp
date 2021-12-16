package fudp

// 可靠文件(夹)传输协议

import (
	"net"

	"github.com/lysShub/fudp/packet"
)

type A interface {
	// CS模式
	Server()
	Client()

	// PP模式
	Send()
	Receive()
}

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
//  通过UDP协议传输文件
func Run(config Config, laddr, raddr string) (err error) {

	return nil
}

func RunWithConn(config Config, conn net.Conn) (err error) {

	return nil
}

/*
	通过默认参数启动的简化函数
*/

// Post 从服务器下载文件
func Pull(url string, path string) (err error) { return }

// Push 上传文件到服务器
func Push(url string, path string) (err error) { return }

// Send 点对点模式发送文件
func Send(path string) (err error) { return }

// Receive 点对点模式接收文件
func Receive(path string) (err error) { return }
