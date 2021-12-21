package fudp

// 可靠文件(夹)传输协议

import (
	"context"
	"net"

	"github.com/lysShub/fudp/packet"
)

type fudp struct {
	Config

	conn      net.Conn   // connected UDP socket
	secretKey [16]byte   // 对称加密(AES_GCM_128)密钥
	gcm       packet.Gcm // 对称加密(AES_GCM_128)
}

// ListenAndServer 必须是CS模式的Server端
func ListenAndServer(addr string, config Config) (err error) {
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
// 	@ca 验签证书, 忽略将使用系统证书
func Pull(ctx context.Context, url string, path string, ca ...[]byte) (err error) { return }

// Push 上传文件/文件夹到服务器
//	@url 请求地址
//	@path 上传的文件/文件夹路径
func Push(ctx context.Context, url string, path string) (err error) { return }

// Send 点对点模式发送文件
//	@path: 被发送文件/文件夹路径
//	@token: 本次传输的token, 接收方需要输入此token
func Send(ctx context.Context, path string, token *string) (err error) { return }

// Receive 点对点模式接收文件
//	@path: 接收文件/文件夹在本地存放路径
// 	@token: 传输token
func Receive(ctx context.Context, path, token string) (err error) { return }
