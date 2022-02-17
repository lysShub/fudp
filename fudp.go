package fudp

import (
	"crypto/cipher"
	"crypto/tls"
	"net"
	"net/url"
)

type fudp struct {
	// fudp 表示一次fudp传输
	// 无论是Client还是Server, 都能上传或

	isP2P    bool // mode
	isClient bool // role

	// 原始connected conn
	rawConn *net.UDPConn

	// 本次传输的URL
	//  url有部分为保留参数, 如act定义此请求是上传还是下载文件
	url *url.URL

	// AES_GCM_128加密密钥, 在握手完成后双方一定被设置
	key [16]byte
	gcm cipher.AEAD

	// Server处理函数, 也可以全局路由注册
	// stateCode 遵从HTTP STATE CODE,
	handleFn Handler

	// 工作路径
	// 提供下载服务时, 是文件在本机磁盘的路径
	// 下载文件时, 是文件存放的路径
	wpath string

	// CS模式基于TLS交换密钥, 仅仅CS模式被设置
	tlsCfg *tls.Config
}

// Pull 从服务器下载文件
func (f *fudp) Pull(path string, url string) {}

// Push 上传文件到服务器
func (f *fudp) Push(path string, url string) {}

func (f *fudp) Close() error {
	return f.rawConn.Close()
}
