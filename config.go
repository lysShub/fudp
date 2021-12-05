package fudp

import (
	"encoding/base32"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/lysShub/fudp/internal/crypter/cert"
	certificate "github.com/lysShub/fudp/internal/crypter/cert"
	"github.com/lysShub/fudp/internal/crypter/ecc"
)

type Config struct {
	mode uint8 // 模式  0: P-P   1: C-S
	role uint8 // 角色  0: Server   1: Client
	acti uint8 // 权限  0b01: 允许下载(receive)  0b10: 允许上传(send) 其他待定

	verifyFunc func(pars *url.URL) bool //请求参数校验函数

	sendPath    string // 发送文件路径, 必须存在, 可以是文件或文件夹
	receivePath string // 接收文件路径, 必须是文件夹

	cert     []byte   // server端的证书
	key      []byte   // server端原始的私钥, 而不是序列化后的
	token    []byte   // PP模式client端一次传输的token, 用于校验证书; 其实质是公钥
	selfCert [][]byte // client端自签根证书

	err error // 配置中遇到的错误
}

func Configure(conf func(*Config)) (Config, error) {
	var c = &Config{}
	conf(c)

	if c.err != nil {
		return Config{}, c.err
	}
	// check

	return *c, nil
}

// ShowToken 显示人类可读的Token(不含不可读字符)
// 	base32编码, 并且将最后的占位符(=)移动至前面, 为了美观
func (c *Config) ShowToken() (token string) {
	if c.mode == 0 && c.role == 0 && len(c.token) > 0 {
		token = base32.StdEncoding.EncodeToString(c.token)

		l, c := len(token), 0
		for i := l - 1; ; i-- {
			if token[i] == '=' {
				c = c + 1
			} else {
				break
			}
		}

		if c == 0 {
			return token
		} else {
			sl := (l - c) / (c + 1)
			var tmpToken string
			for i := 0; i < c; i++ {
				tmpToken = tmpToken + token[i*sl:(i+1)*sl] + "="
			}
			return tmpToken + token[sl*c:l-c]
		}

	}
	return ""
}

const errToken = "0000000000000000"

func (c *Config) ParseToken(token string) []byte {
	if len(token) == 0 {
		return []byte(errToken) // token错误, 会在握手时反馈
	}
	lc := strings.Count(token, "=")
	token = strings.ReplaceAll(token, "=", "")
	for i := 0; i < lc; i++ {
		token = token + "="
	}
	token = strings.ToUpper(token)

	data, err := base32.StdEncoding.DecodeString(token)
	if err != nil {
		return []byte(errToken)
	} else {
		return data
	}
}

// --------------------------------------------------

// --------------------------------------------------

type PPRoler struct {
	*Config
}
type CSRoler struct {
	*Config
}
type CSClient struct {
	*Config
}

// PPMode 点对点模式
func (c *Config) PPMode() iPPMode {
	c.mode = 0
	var pper = PPRoler{c}
	return &pper
}

// CSMode 客户端和服务端模式
func (c *Config) CSMode() iCSMode {
	c.mode = 1
	var cser = CSRoler{c}
	return &cser
}

// Send 发送文件(文件夹)
// 	PP模式下, 发送文件的一方是server端
func (p *PPRoler) Send(path string) {
	var err error
	if err = verifyPath(path, true); err != nil {
		p.err = err
	}
	p.sendPath = formatPath(path)
	p.acti = 0b10
	p.role = 0

	p.cert, p.key, err = certificate.CreateEccCert(nil)
	if err != nil {
		p.err = err
	}

	p.token, _, _, err = certificate.GetCertInfo(p.cert)
	if err != nil {
		p.err = err
	}

	//
	if pub, sign, da, err := cert.GetCertInfo(p.cert); err != nil {
		panic(errors.New("certificate parse fail: " + err.Error()))
	} else {
		fmt.Println("自校验")
		fmt.Println(ecc.Verify(p.token, sign, da))
		fmt.Println("pub", pub)
		fmt.Println("token", p.token)
	}
}

// Receive 接收文件(文件夹)
func (p *PPRoler) Receive(path string, token []byte) {
	var err error
	if err = verifyPath(path, false); err != nil {
		p.err = err
	}

	p.receivePath = formatPath(path)
	p.acti = 0b1
	p.role = 1

	// check token

	p.token = token
}

// Client 客户端
// 	@rootCertificate: 验签使用的根证书, 不填写将使用系统CA根证书
func (c *CSRoler) Client(rootCertificate ...[]byte) iCSClient {
	var csclient = CSClient{c.Config}
	csclient.selfCert = rootCertificate
	for i, v := range rootCertificate {
		if !certificate.CertFormatCheck(v) {
			c.err = errors.New("rootCertificate format error, index " + strconv.Itoa(i))
		}
	}
	c.role = 1
	return &csclient
}

// Server 服务端
func (c *CSRoler) Server(cert []byte, key []byte, verifyFunc func(pars *url.URL) uint8) iCSServer {
	var csserver = CSServer{c.Config}
	csserver.cert = cert
	csserver.key = key
	if !certificate.CertFormatCheck(cert) {
		c.err = errors.New("cert format error")
	}
	c.role = 0
	return &csserver
}

// Send 客户端发送文件(文件夹)
func (c *CSClient) Send(path string) {
	var err error
	if err = verifyPath(path, true); err != nil {
		c.err = err
	}

	c.sendPath = formatPath(path)
	c.acti = 0b10
}

// Receiver 客户端接收文件(文件夹)
func (c *CSClient) Receive(path string) {
	var err error
	if err = verifyPath(path, false); err != nil {
		c.err = err
	}
	c.receivePath = formatPath(path)
	c.acti = 0b1
}

type CSServer struct {
	*Config
}

// Send 服务端发送文件(文件夹)
func (c *CSServer) Send(path string) {
	c.sendPath = path
	c.acti = 0b10
}

// Receive 服务端接收文件(文件夹)
func (c *CSServer) Receive(path string) {
	c.receivePath = path
	c.acti = 0b1
}

// All 服务端发送/接收文件(文件夹)
func (c *CSServer) All(spath string, rpath string) {
	c.sendPath = spath
	c.receivePath = rpath
	c.acti = 0b11
}

// --------------------------------------------------

// --------------------------------------------------
type iConf interface {
	PPMode() iPPMode
	CSMode() iCSMode
}

type iPPMode interface {
	Send(path string)
	Receive(path string, token []byte)
}
type iCSMode interface {
	Client(rootCertificate ...[]byte) iCSClient
	Server(cert []byte, key []byte, verifyFunc func(pars *url.URL) uint8) iCSServer
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
