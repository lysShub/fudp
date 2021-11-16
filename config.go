package fudp

import (
	"errors"
	"strconv"

	certificate "github.com/lysShub/fudp/internal/crypter/cert"
)

// New 创建一个新的fudp实例
func New(conf func(*Configure)) *Configure {
	var f = new(fudp)
	var c = &Configure{f}
	conf(c)

	// check

	return c
}

// --------------------------------------------------
type iConf interface {
	PPMode() iPPMode
	CSMode() iCSMode
}
type iPPMode interface {
	Send(path string) (token []byte)
	Receive(path string, token []byte)
}
type iCSMode interface {
	Client(rootCertificate ...[]byte) iCSClient
	Server(cert []byte, key []byte) iCSServer
}

type iCSClient interface {
	Send(path string)
	Receive(path string)
}

type iCSServer interface {
	Send(path string)
	Receive(path string)
	All(spath string, rpath string)
}

// --------------------------------------------------

type Configure struct {
	*fudp
}

// PPMode 点对点模式
func (c *Configure) PPMode() iPPMode {
	c.mode = 0
	var pper = PPRoler{c}
	return &pper
}

// CSMode 客户端和服务端模式
func (c *Configure) CSMode() iCSMode {
	c.mode = 1
	var cser = CSRoler{c}
	return &cser
}

type PPRoler struct {
	*Configure
}
type CSRoler struct {
	*Configure
}

// Send 发送文件(文件夹)
func (p *PPRoler) Send(path string) (token []byte) {
	var err error
	if err = verifyPath(path, true); err != nil {
		p.err = err
	}
	p.sendPath = formatPath(path)
	p.auth = 0b10

	p.cert, p.key, err = certificate.CreateEccCert(nil)
	if err != nil {
		p.err = err
	}

	token, _, _, err = certificate.GetCertInfo(p.cert)
	if err != nil {
		p.err = err
	}
	return
}

// Receive 接收文件(文件夹)
func (p *PPRoler) Receive(path string, token []byte) {
	var err error
	if err = verifyPath(path, true); err != nil {
		p.err = err
	}

	p.receivePath = formatPath(path)
	p.auth = 0b1

	// check token
	p.tocken = token
}

// Client 客户端
// 	@rootCertificate: 验签使用的根证书, 不填写将使用系统CA根证书
func (c *CSRoler) Client(rootCertificate ...[]byte) iCSClient {
	var csclient = CSClient{c.Configure}
	csclient.selfCert = rootCertificate
	for i, v := range rootCertificate {
		if !certificate.CertFormatCheck(v) {
			c.err = errors.New("rootCertificate format error, index " + strconv.Itoa(i))
		}
	}

	return &csclient
}

// Server 服务端
func (c *CSRoler) Server(cert []byte, key []byte) iCSServer {
	var csserver = CSServer{c.Configure}
	csserver.cert = cert
	csserver.key = key
	if !certificate.CertFormatCheck(cert) {
		c.err = errors.New("cert format error")
	}

	return &csserver
}

type CSClient struct {
	*Configure
}

// Send 客户端发送文件(文件夹)
func (c *CSClient) Send(path string) {
	c.sendPath = path
	c.auth = 0b10
}

// Receiver 客户端接收文件(文件夹)
func (c *CSClient) Receive(path string) {
	c.receivePath = path
	c.auth = 0b1
}

type CSServer struct {
	*Configure
}

// Send 服务端发送文件(文件夹)
func (c *CSServer) Send(path string) {
	c.sendPath = path
	c.auth = 0b10
}

// Receive 服务端接收文件(文件夹)
func (c *CSServer) Receive(path string) {
	c.receivePath = path
	c.auth = 0b1
}

// All 服务端发送/接收文件(文件夹)
func (c *CSServer) All(spath string, rpath string) {
	c.sendPath = spath
	c.receivePath = rpath
	c.auth = 0b11
}

// --------------------------------------------------
