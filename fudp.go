package fudp

import (
	"crypto/cipher"
	"crypto/tls"
	"net"
	"net/url"
)

type fudp struct {
	rawConn *net.UDPConn

	url url.URL

	// P2P模式必须设置, CS必须为空值; 与tlsConfig互斥
	key [16]byte

	// P2P模式为nil, 否则为CS模式
	tlsConfig *tls.Config

	gcm cipher.AEAD
}
