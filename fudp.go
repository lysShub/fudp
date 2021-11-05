package fudp

import (
	"net"
	"os"
)

// 可靠文件(夹)传输协议

type fudp struct {
	// 一个fudp代表一次传输

	// AES_GCM 对称加密
	secretKey    [16]byte                          // 加密密钥
	none         [12]byte                          // 随机值
	secretKeySet bool                              // 是否设置加密
	SetSecretKey func(key [16]byte, none [12]byte) // 设置密钥
}

type Conf struct {
	// 实例启动参数

	// 实例启动的模式
	// 比如启动一个Listen实例, 这个实例可以存在至少三种模式：只允许下载、只许上传、支持上传下载
	// 第一位为1: 下载; 第二位为1: 上传; .....待完善
	mode uint8

	// 上传, 下载根目录
	// 上传根目录指传给对方文件所在本机的根目录
	// 下载根目录指对方发过来的文件存储在本机的根目录
	uPath, dPath string

	// 公钥，在传输双方供中接收方使用。如果不为nil，验签时将最优先使用; 如果不填写将依次尝试
	// CA证书、自签证书验签。
	// 更多信息查阅：
	// https://github.com/lysShub/fudp
	publicKey []byte
}

func NewUploadConf(path string, f func(*Conf) *Conf) (*Conf, error) {
	os.Stat(path)

	var c = new(Conf)
	c = f(c)
	c.mode = 0b10

	return c, nil
}

func NewDownloadConf(path string, f func(*Conf) *Conf) (*Conf, error) {
	os.Stat(path)

	var c = new(Conf)
	c = f(c)
	c.mode = 0b01
	return c, nil
}

func (f *fudp) Listen(laddr *net.UDPAddr, conf *Conf) error {
	return nil
}

func (f *fudp) UpLoad(raddr *net.UDPAddr, conf *Conf) error {
	return nil
}
