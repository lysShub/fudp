package fudp

// 可靠文件(夹)传输协议
// 点对点、端到端

import (
	"net"
	"net/url"
)

// fudp的工作模式
// 点对点、端到端

type fudp struct {
	// 一个fudp代表一次传输
	// 网络
	// connected sokcet; P-P模式时双方都使用dial创建原始的udp connect
	// S-C模式时, Client使用dial创建原始的udp connect, Server使用ioer
	conn net.Conn // connected socket

	// 工作模式
	// 0: P-P   1: C-S
	mode uint8

	// 权限
	// 0b1: 允许下载(receive)  0b10: 允许上传(send) 其他待定
	auth          uint8
	UserVerifyAct func(pars *url.URL) uint8 // 用户握手请求校验, 握手包0中的格式化数据

	// 上传下载路径
	sendPath, receivePath string

	/* 安全 */
	cert      []byte   // 接受握手方的证书
	key       []byte   // 接受握手方的非对称加密的私钥
	tocken    []byte   // 安全令牌; P-P模式时Client使用, 用于校验证书; 其实是证书的公钥
	selfCert  [][]byte // 自签根证书
	secretKey [16]byte // 对称加密(AES_GCM_128)密钥

	err error // 配置中遇到的错误
}

// Run 启动
func (f *fudp) Run() error {
	return nil
}
