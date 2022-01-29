package sconn_test

import (
	"crypto/tls"
	"net"
	"testing"
	"time"

	"github.com/lysShub/fudp/internal/sconn"
	"github.com/stretchr/testify/require"
)

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

func TestSconn(t *testing.T) {

	go func() {
		cert, err := tls.X509KeyPair([]byte(cert), []byte(key))
		require.Nil(t, err)

		conn, err := net.DialUDP("udp", &net.UDPAddr{Port: 19986}, &net.UDPAddr{Port: 19987})
		require.Nil(t, err)
		sconn := sconn.NewSconn(conn)
		tconn := tls.Server(sconn, &tls.Config{Certificates: []tls.Certificate{cert}})
		defer tconn.Close()

		err = tconn.Handshake()
		require.Nil(t, err)

		handleFunc(tconn, t) // echo server
	}()

	time.Sleep(time.Second)
	client(t)
}

func handleFunc(conn net.Conn, t *testing.T) {
	defer conn.Close()

	var buf []byte = make([]byte, 2000)
	for {
		if n, err := conn.Read(buf); err != nil {
			panic(err)
		} else {
			if n == 0 {
				t.Fatal(n)
			} else {
				m, err := conn.Write(buf[:n])
				require.Nil(t, err)
				require.Equal(t, n, m)
			}
		}
	}
}

func client(t *testing.T) {

	conn, err := net.DialUDP("udp", &net.UDPAddr{Port: 19987}, &net.UDPAddr{Port: 19986})
	require.Nil(t, err)
	sconn := sconn.NewSconn(conn)

	tconn := tls.Client(sconn, &tls.Config{
		InsecureSkipVerify: true,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
		},
	})

	err = tconn.Handshake()
	require.Nil(t, err)

	time.Sleep(time.Second)

	// hello world
	n, err := tconn.Write([]byte("hello world!"))
	require.Equal(t, 12, n)
	require.Nil(t, err)
	var da = make([]byte, 16)
	n, err = tconn.Read(da)
	require.Equal(t, 12, n)
	require.Nil(t, err)
	require.Equal(t, []byte("hello world!"), da[:n])

}