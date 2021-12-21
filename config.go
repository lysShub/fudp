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

	"github.com/lysShub/fudp/internal/crypter/cert"
	"github.com/lysShub/fudp/internal/crypter/ecc"
	"github.com/lysShub/fudp/utils"
)

type mode uint8

const (
	PPMode mode = 1 << iota
	CSMode
)

type role uint8

const (
	SRole role = 1 << iota // server role: 接受握手
	CRole
)

type act uint8

const (
	DownloadAct act = 1 << iota
	UploadAct
)

type Config struct {
	mode        mode   // 模式
	role        role   // 角色
	acti        act    // 权限
	sendPath    string // 发送文件路径, 支持文件、文件夹
	receivePath string // 接收文件路径

	cert     []byte            // 证书
	key      *ecdsa.PrivateKey // 私钥
	token    []byte            // token, 用于校验证书; 等同于CS模式中证书的公钥
	selfCert [][]byte          // 自签根证书的验签证书

	url        string                                       // C端url: fudp://host:port/download?token=xxx&systen=windows
	handleFunc func(pars *url.URL) (path string, err error) // 处理请求, 返回本地路径
	err        error                                        // 配置错误
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
	c.mode = PPMode
	var pper = PPRoler{c}
	return &pper
}

func (c *Config) CSMode() iCSMode {
	c.mode = CSMode
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
	p.acti = UploadAct
	p.role = SRole

	p.cert, p.key, err = cert.GenerateCert(time.Hour*24, func(c *cert.CaRequest) {})
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

// Receive 接收文件(文件夹)
// @url: 请求地址、参数
// @path: 接收文件本地存放路径
func (p *PPRoler) Receive(url string, path string) {
	p.url = url

	var err error
	if err = verifyPath(path, false); err != nil {
		p.err = err
		return
	}

	p.receivePath = utils.FormatPath(path)
	p.acti = DownloadAct
	p.role = CRole

	p.key, err = ecc.GenerateKey()
	if err != nil {
		p.err = err
		return
	}

	pk, err := ecc.MarshalPubKey(&p.key.PublicKey)
	if err != nil {
		p.err = err
		return
	}
	p.token = pk
}

// Client 客户端
// 	@rootCertificate: 验签使用的根证书, 不填写将使用系统CA根证书
func (c *CSRoler) Client(rootCertificate ...[]byte) {
	var csclient = CSClient{c.Config}
	csclient.selfCert = rootCertificate
	for i, v := range rootCertificate {
		if !cert.CertFormatCheck(v) {
			c.err = errors.New("rootCertificate format error, index " + strconv.Itoa(i))
		}
	}
	c.role = CRole
}

// Server 服务端
//	@cert 证书
//	@key 私钥
//	@verifyFunc url参数校验函数, 可以作为请求的Auth, nil表示校验成功, 否则回复401, 且err作为回复信息(明文)
func (c *CSRoler) Server(cert []byte, key []byte, handleFunc func(pars *url.URL) (path string, err error)) {
	var csserver = CSServer{c.Config}
	csserver.cert = cert
	var err error
	if csserver.key, err = ecc.ParsePriKey(key); err != nil {
		c.err = err
	} else {
		c.role = 0
	}

	// 处理验证保留参数
	c.handleFunc = handleFunc
}

// ShowToken 显示序列化后的token
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
	Receive(url string, path string)
}
type iCSMode interface {
	Client(rootCertificate ...[]byte)

	// Server
	//	@cert: 证书
	//	@key: 密钥
	//	@handleFunc：处理函数
	Server(cert []byte, key []byte, handleFunc func(pars *url.URL) (path string, err error))
}

// --------------------------------------------------

// --------------------------------------------------

// 未被使用
func (c *Config) parseToken(token string) []byte {
	if len(token) == 0 {
		return make([]byte, 16) // token错误, 会在握手时反馈
	}
	lc := strings.Count(token, "=")
	token = strings.ReplaceAll(token, "=", "")
	for i := 0; i < lc; i++ {
		token = token + "="
	}
	token = strings.ToUpper(token)

	data, err := base32.StdEncoding.DecodeString(token)
	if err != nil {
		return make([]byte, 16)
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
