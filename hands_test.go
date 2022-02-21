package fudp

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHand_P2P(t *testing.T) {

	// 测试P2P握手
	// 发送端开放19986端口, 接收端开放19987端口
	rconn, err := net.DialUDP("udp", &net.UDPAddr{Port: 19987}, &net.UDPAddr{Port: 19986})
	require.NoError(t, err)
	defer rconn.Close()
	sconn, err := net.DialUDP("udp", &net.UDPAddr{Port: 19986}, &net.UDPAddr{Port: 19987})
	require.NoError(t, err)
	defer sconn.Close()

	u, err := url.Parse("fudp://localhost:19986/a/bb?mtu=1372")
	require.NoError(t, err)
	var si = &fudp{
		isP2P:    true,
		isClient: false,
		rawConn:  sconn,
		key:      [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		handleFn: func(url *url.URL) (path string, stateCode int) { return "", 200 },
	}
	var ci = &fudp{
		isP2P:    true,
		isClient: true,
		rawConn:  rconn,
		key:      [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		url:      u,
	}

	// 发送方
	go func() {
		require.NoError(t, si.handPong())
	}()
	statueCode, err := ci.handPing()
	require.NoError(t, err)
	require.Equal(t, uint16(200), statueCode)
}

func TestHand_CS(t *testing.T) {

	// 测试CS握手
	// 发送端开放19986端口, 接收端开放19987端口

	var cert string = `-----BEGIN CERTIFICATE-----
MIIBbjCCARSgAwIBAgIRAI+jBYEYS5aBXDUedBt7PKYwCgYIKoZIzj0EAwIwEjEQ
MA4GA1UEChMHQWNtZSBDbzAeFw0yMjAxMDkxNzQxMjBaFw0yMzAxMDkxNzQxMjBa
MBIxEDAOBgNVBAoTB0FjbWUgQ28wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQy
l1/gWhTxZ3rS/XJOMLHhmkQp64EtPrEgq9SjKDpWBZQC+kNZdM5xzJrv3bLqcyOS
JywZfEpTZzW7sxko4maBo0swSTAOBgNVHQ8BAf8EBAMCB4AwEwYDVR0lBAwwCgYI
KwYBBQUHAwEwDAYDVR0TAQH/BAIwADAUBgNVHREEDTALgglsb2NhbGhvc3QwCgYI
KoZIzj0EAwIDSAAwRQIhAICxMC8o603GwL3bf42EXrtPP5/LtEIc/hjdJpilqc3b
AiBTEdrE+/oCgUjsxV2RFj1+42CTGtcav4sJyCPjme0N/w==
-----END CERTIFICATE-----`

	var key string = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgmH+BleetLN1fK0dy
JpedWG8C2yxtb7gEEAwvdwXf6FihRANCAAQyl1/gWhTxZ3rS/XJOMLHhmkQp64Et
PrEgq9SjKDpWBZQC+kNZdM5xzJrv3bLqcyOSJywZfEpTZzW7sxko4maB
-----END PRIVATE KEY-----`

	rconn, err := net.DialUDP("udp", &net.UDPAddr{Port: 19987}, &net.UDPAddr{Port: 19986})
	require.NoError(t, err)
	defer rconn.Close()
	sconn, err := net.DialUDP("udp", &net.UDPAddr{Port: 19986}, &net.UDPAddr{Port: 19987})
	require.NoError(t, err)
	defer sconn.Close()

	c, err := tls.X509KeyPair([]byte(cert), []byte(key))
	require.NoError(t, err)
	tlsCfg := &tls.Config{ClientAuth: tls.NoClientCert, Certificates: []tls.Certificate{c}}

	p, _ := pem.Decode([]byte(cert))
	cc, err := x509.ParseCertificate(p.Bytes)
	require.NoError(t, err)

	u, err := url.Parse("fudp://localhost:19986/a/bb?mtu=1372")
	require.NoError(t, err)
	var si = &fudp{
		isP2P:    false,
		isClient: false,
		rawConn:  sconn,
		handleFn: func(url *url.URL) (path string, stateCode int) { return "", 200 },
		tlsCfg:   tlsCfg,
	}
	var ci = &fudp{
		isP2P:    false,
		isClient: true,
		rawConn:  rconn,
		url:      u,
	}

	go func() {
		require.NoError(t, si.handPong())
	}()

	statueCode, err := ci.handPing(cc)
	require.NoError(t, err)
	require.Equal(t, uint16(200), statueCode)
	// time.Sleep(time.Second * 5)
}
