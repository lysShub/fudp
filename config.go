package fudp

import (
	"crypto/ecdsa"
	"encoding/base32"
	"errors"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	certificate "github.com/lysShub/fudp/internal/crypter/cert"
	"github.com/lysShub/fudp/internal/crypter/ecc"
	"github.com/lysShub/fudp/utils"
)

type Config struct {
	mode uint8 // 模式  0: P-P   1: C-S
	role uint8 // 角色  0: Server   1: Client
	acti uint8 // 权限  0b01: 允许下载(receive)  0b10: 允许上传(send) 其他待定

	verifyFunc func(pars *url.URL) bool //请求参数校验函数

	sendPath    string // 发送文件路径, 必须存在, 可以是文件或文件夹
	receivePath string // 接收文件路径, 必须是文件夹

	cert     []byte            // server端的证书
	key      *ecdsa.PrivateKey // server端私钥
	token    []byte            // PP模式client端一次传输的token, 用于校验证书; 其实质是公钥
	selfCert [][]byte          // client端自签根证书

	err error // 配置中遇到的错误
}

// Configure 创建配置文件, 具有向导功能
func Configure(conf func(*Config)) (Config, error) {
	var c = &Config{}
	conf(c)

	if c.err != nil {
		return Config{}, c.err
	}
	// check

	return *c, nil
}

func (c *Config) PPMode() iPPMode {
	c.mode = 0
	var pper = PPRoler{c}
	return &pper
}

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
	p.sendPath = utils.FormatPath(path)
	p.acti = 0b10
	p.role = 0

	p.cert, p.key, err = certificate.GenerateCert(time.Hour*24, func(c *certificate.CaRequest) {})
	if err != nil {
		p.err = err
		return
	}

	p.token, err = ecc.MarshalPubKey(&p.key.PublicKey)
	if err != nil {
		p.err = err
		return
	}
}

// Send 客户端发送文件(文件夹)
func (c *CSClient) Send(path string) {
	var err error
	if err = verifyPath(path, true); err != nil {
		c.err = err
	}

	c.sendPath = utils.FormatPath(path)
	c.acti = 0b10
}

// Receive 接收文件(文件夹)
func (p *PPRoler) Receive(path string, token string) {
	var err error
	if err = verifyPath(path, false); err != nil {
		p.err = err
	}

	p.receivePath = utils.FormatPath(path)
	p.acti = 0b1
	p.role = 1

	p.token = p.parseToken(token)
}

// Receiver 客户端接收文件(文件夹)
func (c *CSClient) Receive(path string) {
	var err error
	if err = verifyPath(path, false); err != nil {
		c.err = err
	}
	c.receivePath = utils.FormatPath(path)
	c.acti = 0b1
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
// 	@cert 证书
// 	@key 私钥
//  @verifyFunc 对请求参数进行校验
func (c *CSRoler) Server(cert []byte, key []byte, verifyFunc func(pars *url.URL) uint8) iCSServer {
	var csserver = CSServer{c.Config}
	csserver.cert = cert
	var err error
	if csserver.key, err = ecc.ParsePriKey(key); err != nil {
		c.err = err
	} else {
		if !certificate.CertFormatCheck(cert) { // 检查证书格式, 有效期, 公私钥匹配
			c.err = errors.New("cert format error")
		}
		c.role = 0
	}
	return &csserver
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

type CSServer struct {
	*Config
}
type PPRoler struct {
	*Config
}
type CSRoler struct {
	*Config
}
type CSClient struct {
	*Config
}
type iPPMode interface {
	Send(path string)
	Receive(path string, token string)
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

// --------------------------------------------------

const errToken = "0000000000000000"

func (c *Config) parseToken(token string) []byte {
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

func verifyPath(path string, isSend bool) error {
	if isSend {
		fi, err := os.Stat(path)
		if os.IsNotExist(err) {
			return errors.New("invalid path: not exist")
		} else {
			if !fi.IsDir() {
				if fi.Size() == 0 {
					return errors.New("invalid path: file empty")
				}
			} else {
				var s int64
				filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
					s = s + info.Size()
					if s > 0 {
						return errors.New("null")
					}
					return nil
				})
				if s == 0 {
					return errors.New("invalid path: path empty")
				}
			}
		}
	} else {
		fi, err := os.Stat(path)
		if os.IsNotExist(err) {
			return os.MkdirAll(path, 0666)
		} else if !fi.IsDir() {
			return errors.New("invalid path: is file path, expcet floder path")
		}
	}
	return nil
}
