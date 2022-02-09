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

	rawConn *net.UDPConn

	url *url.URL

	// P2P模式必须设置, CS必须为空值; 与tlsConfig互斥
	key [16]byte

	// P2P模式为nil, 否则为CS模式
	cfg *tls.Config

	// 传输加密AES_GCM实例
	gcm cipher.AEAD

	// Server处理函数, 也可以全局路由注册
	// stateCode 遵从HTTP STATE CODE,
	// msg可选, 将被返回给client
	handlFn func(url *url.URL) (path string, stateCode int)

	// 工作路径
	path string
}

var pType = struct {
	handPackage int
}{
	0,
}
