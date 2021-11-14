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
	conn *net.UDPConn // connected socket

	// 传输
	mut int

	// 工作模式
	// 0: P-P   1: C-S
	mode uint8

	// 权限
	// 0b1: 允许下载  0b10: 允许上传 其他待定
	auth          uint8
	UserVerifyAct func(pars *url.URL) uint8 // 用户握手请求校验, 握手包0中的格式化数据

	// 上传下载路径
	upLoadDir, downLoadDir string

	/* 安全 */
	caCert, caKey []byte   // 证书、密钥 (ECC 256)。对于Server两个都必须存在；对于Client证书存在代表是自签证书
	pubKey        []byte   // 非对称加密公钥。 对于Server相对于证书; 对于Client则是验签的公钥用于验签
	secretKey     [16]byte // 对称加密(AES_GCM_128)密钥
}

// Run 启动
func (f *fudp) Run() error {

	return nil
}
