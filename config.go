package fudp

import (
	"crypto/ecdsa"
	"encoding/base32"
	"errors"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lysShub/fudp/internal/crypter/cert"
	"github.com/lysShub/fudp/internal/crypter/ecc"
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

type Config struct {
	mode     mode   // 模式
	role     role   // 角色
	rootPath string // 根路径, 接收文件则是存放路径, 发送文件这是发送文件根路径

	cert     []byte            // 证书
	key      *ecdsa.PrivateKey // 私钥
	token    []byte            // token, 用于校验证书; 等同于CS模式中证书的公钥
	selfCert [][]byte          // 自签根证书的验签证书

	url        string  // C端url: fudp://host:port/download?token=xxx&systen=windows
	handleFunc Handler // 不设置使用全局注册handle, 设置只使用此函数
	err        error   // 配置错误
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

// PPMode 点对点模式
func (c *Config) PPMode() iPPMode {
	c.mode = PPMode
	var pper = PPRoler{c}
	return &pper
}

// CSMode client-server模式
func (c *Config) CSMode() iCSMode {
	c.mode = CSMode
	var cser = CSRoler{c}
	return &cser
}

// Send 发送文件(文件夹)
// 	PP模式下, 规定发送文件的一方是server端
//	@path: 发送文件的路径gn
//	@ignore: 是否忽略不可读的文件, 如果设置为fals, 将会遍历所有文件
func (p *PPRoler) Send(path string, ignore bool) {
	var err error
	if p.rootPath, err = formatPath(path); err != nil {
		p.err = err
		return
	}
	if p.err = checkSendPath(path, ignore); p.err != nil {
		return
	}

	p.role = SRole
	// 处理函数
	p.handleFunc = func(url *url.URL) (path string, stateCode int, msg string) {
		if url.Path, err = formatPath(url.Path); err != nil {
			return "", 0, err.Error()
		}
		// rootPath经过Abs处理, 不会存在父路径
		if reqPath := filepath.Join(p.rootPath, url.Path); len(reqPath) > len(p.rootPath) {
			if _, err := os.Stat(reqPath); err == nil {
				return reqPath, http.StatusOK, "ok"
			} else if os.IsNotExist(err) {
				return "", http.StatusNotFound, "not found"
			}
		}
		return "", http.StatusBadRequest, "invalid requset"
	}
}

// Receive 接收文件(文件夹)
// @url: 请求地址、参数
// @path: 接收文件本地存放路径
func (p *PPRoler) Receive(path string) {
	if p.rootPath, p.err = formatPath(path); p.err != nil {
		return
	}
	if fi, err := os.Stat(p.rootPath); os.IsNotExist(err) {
		p.err = errors.New("'" + p.rootPath + "' not exist")
		return
	} else if !fi.IsDir() {
		p.err = errors.New("'" + p.rootPath + "' expect floder, got file path")
		return
	}

	p.role = CRole
}

// Client 客户端
// 	@rootCertificate: 验签使用的根证书, 不填写将使用系统CA根证书
func (c *CSRoler) Client(rootSelfCertificate ...[]byte) {
	var csclient = CSClient{c.Config}
	csclient.selfCert = rootSelfCertificate
	for i, v := range rootSelfCertificate {
		if !cert.CheckCertFormat(v) {
			c.err = errors.New("rootCertificate format error, index " + strconv.Itoa(i))
		}
	}
	c.role = CRole
}

// Server 服务端
//	@cert 证书
//	@key 私钥
//	@handleFunc 处理函数, 返回本地路径
func (c *CSRoler) Server(cert []byte, key []byte, handleFunc Handler) {
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
	// Send 发送文件
	//	@path 发送文件本地路径
	// PP模式规定发送方是server, 接收方是client
	Send(path string, ignore bool)

	// Receive 发送文件
	//	@path 发送文件本地路径
	// PP模式规定发送方是server, 接收方是client
	Receive(path string)
}
type iCSMode interface {

	// Client
	//
	Client(rootCertificate ...[]byte)

	// Server
	//	@cert: 证书
	//	@key: 密钥
	//	@handleFunc：处理函数
	Server(cert []byte, key []byte, handleFunc Handler)
}

// --------------------------------------------------

// --------------------------------------------------

func formatPath(path string) (string, error) {
	switch os.PathSeparator {
	case '\\':
		path = filepath.FromSlash(path)
	case '/':
		path = filepath.ToSlash(path)
	default:
		path = ""
	}
	return filepath.Abs(path)
}

func checkSendPath(path string, ignore bool) (err error) {
	if fi, e := os.Stat(path); os.IsNotExist(e) {
		return errors.New("'" + path + "' not exist")
	} else if fi.IsDir() {
		var ts int
		filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
			if !info.IsDir() && err == nil {
				if !ignore { // 判断是否可读
					if fh, err := os.Open(path); err != nil {
						err = errors.New("'" + path + "' can't read: " + err.Error())
						return errors.New("null") // 退出
					} else {
						defer fh.Close()
					}
				}
				ts = ts + int(fi.Size())
			}
			if ignore && ts > 0 {
				return errors.New("null") // 退出
			}
			return nil
		})
		if ts <= 0 {
			return errors.New("'" + path + "' is empty")
		}
	} else {
		if fh, e := os.Open(path); e != nil {
			err = errors.New("'" + path + "' can't read: " + err.Error())
			return
		} else {
			defer fh.Close()
			if fi.Size() == 0 {
				return errors.New("'" + path + "' is empty file")
			}
		}
	}
	return nil
}

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
