package constant

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"
)

// 定量

// 协议包最大的payload大小, 单位字节
// 协议包头大小为[3,13]字节
const MTU = 5000

// 默认等待握手开始超时时间
const WaitHandShakeTimeout = time.Second * 4

// 握手开始后, 整个握手过程的超时时间
// 理论耗时：C端1.5RTT, S端1RTT
const HandshakeTimeout = time.Millisecond * 400

// S端参数, Client允许的最大重试握手次数
const HandshakeRetry = 3

// 证书模板
var CertTemplate = x509.Certificate{
	SerialNumber: big.NewInt(int64(0)), // CA颁发证书对应的唯一序列号，自签填个随机数即可
	Subject: pkix.Name{
		Organization: []string{"Acme Co"}, // 机构
	},
	NotBefore:             time.Now(),
	NotAfter:              time.Now().AddDate(0, 0, 1),                           // 有效期结束时间
	KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign, // 使用场景
	ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	BasicConstraintsValid: true,
}