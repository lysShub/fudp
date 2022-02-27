package constant

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"
)

// 定量

const MTU = 65536                      // fudp协议数据包最大大小（包括包头）
const HeadSize = 9                     // fudp协议数据包头大小
const MaxPayload = MTU - HeadSize - 16 // fudp协议数据包最大承载数据大小; 16为AES加密增加的数据

// 握手开始后, 整个握手过程的超时时间
// 理论耗时：C端1.5RTT, S端1RTT
const HandshakeTimeout time.Duration = time.Millisecond * 400

// RTT 数据包往返时间, 及PING的时间
const RTT = time.Millisecond * 250

// 加密位数
const SIZE = 16

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
