package utls_test

import (
	"crypto/tls"
	"crypto/x509"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/lysShub/fudp/internal/utls"
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

var invalidCert string = `-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----`

var key string = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgmH+BleetLN1fK0dy
JpedWG8C2yxtb7gEEAwvdwXf6FihRANCAAQyl1/gWhTxZ3rS/XJOMLHhmkQp64Et
PrEgq9SjKDpWBZQC+kNZdM5xzJrv3bLqcyOSJywZfEpTZzW7sxko4maB
-----END PRIVATE KEY-----`

func TestSconn_Base(t *testing.T) {
	// echo server
	go func() {
		cert, err := tls.X509KeyPair([]byte(cert), []byte(key))
		require.NoError(t, err)

		conn, err := net.DialUDP("udp", &net.UDPAddr{Port: 19986}, &net.UDPAddr{Port: 19987})
		require.NoError(t, err)
		sconn := utls.NewSconn(conn)
		tconn := tls.Server(sconn, &tls.Config{ClientAuth: tls.NoClientCert, Certificates: []tls.Certificate{cert}})
		defer tconn.Close()

		require.NoError(t, tconn.SetReadDeadline(time.Now().Add(time.Second*2)))
		require.NoError(t, tconn.Handshake())
		require.NoError(t, tconn.SetReadDeadline(time.Time{}))

		var buf []byte = make([]byte, 2000)
		for {
			if n, err := tconn.Read(buf); err != nil {
				panic(err)
			} else {
				if n == 0 {
					t.Fatal(n)
				} else {
					m, err := tconn.Write(buf[:n])
					require.NoError(t, err)
					require.Equal(t, n, m)
				}
			}
		}
	}()

	time.Sleep(time.Second)

	// client
	conn, err := net.DialUDP("udp", &net.UDPAddr{Port: 19987}, &net.UDPAddr{Port: 19986})
	require.NoError(t, err)
	sconn := utls.NewSconn(conn)

	p := x509.NewCertPool()
	require.Equal(t, true, p.AppendCertsFromPEM([]byte(cert)))
	tconn := tls.Client(sconn, &tls.Config{
		InsecureSkipVerify: false,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
		},
		ServerName: "localhost",
		RootCAs:    p,
	})

	require.NoError(t, tconn.SetReadDeadline(time.Now().Add(time.Second)))
	require.NoError(t, tconn.Handshake())
	require.NoError(t, tconn.SetReadDeadline(time.Time{}))

	// hello world
	n, err := tconn.Write([]byte("hello world!"))
	require.NoError(t, err)
	require.Equal(t, 12, n)

	var da = make([]byte, 36)
	n, err = tconn.Read(da)
	require.NoError(t, err)
	require.Equal(t, 12, n)
	require.Equal(t, []byte("hello world!"), da[:n])
}

func TestSconn_InvalidCert(t *testing.T) {
	// echo server
	go func() {
		cert, err := tls.X509KeyPair([]byte(cert), []byte(key))
		require.NoError(t, err)

		conn, err := net.DialUDP("udp", &net.UDPAddr{Port: 19986}, &net.UDPAddr{Port: 19987})
		require.NoError(t, err)
		sconn := utls.NewSconn(conn)
		tconn := tls.Server(sconn, &tls.Config{ClientAuth: tls.NoClientCert, Certificates: []tls.Certificate{cert}})
		defer tconn.Close()

		require.NoError(t, tconn.SetReadDeadline(time.Now().Add(time.Second*2)))
		require.NotNil(t, tconn.Handshake())
	}()

	time.Sleep(time.Second)

	// client
	conn, err := net.DialUDP("udp", &net.UDPAddr{Port: 19987}, &net.UDPAddr{Port: 19986})
	require.NoError(t, err)
	sconn := utls.NewSconn(conn)
	defer sconn.Close()

	p := x509.NewCertPool()
	require.Equal(t, true, p.AppendCertsFromPEM([]byte(invalidCert)))
	tconn := tls.Client(sconn, &tls.Config{
		InsecureSkipVerify: false,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
		},
		ServerName: "localhost",
		RootCAs:    p,
	})

	require.NoError(t, tconn.SetReadDeadline(time.Now().Add(time.Second)))
	err = tconn.Handshake()
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "authority")
	// x509: certificate signed by unknown authority (possibly because of "x509: ECDSA verification failure" while trying to verify candidate authority certificate "Acme Co")
}

func TestSconn_Timeout(t *testing.T) {
	// server
	cert, err := tls.X509KeyPair([]byte(cert), []byte(key))
	require.NoError(t, err)

	conn, err := net.DialUDP("udp", &net.UDPAddr{Port: 19986}, &net.UDPAddr{Port: 19987})
	require.NoError(t, err)
	sconn := utls.NewSconn(conn)
	tconn := tls.Server(sconn, &tls.Config{ClientAuth: tls.NoClientCert, Certificates: []tls.Certificate{cert}})

	require.NoError(t, tconn.SetReadDeadline(time.Now().Add(time.Millisecond*900)))
	require.Contains(t, tconn.Handshake().Error(), "timeout")
	tconn.Close()

	// client
	conn, err = net.DialUDP("udp", &net.UDPAddr{Port: 19987}, &net.UDPAddr{Port: 19986})
	require.NoError(t, err)
	sconn = utls.NewSconn(conn)

	p := x509.NewCertPool()
	require.Equal(t, true, p.AppendCertsFromPEM([]byte(invalidCert)))
	tconn = tls.Client(sconn, &tls.Config{
		InsecureSkipVerify: false,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
		},
		ServerName: "localhost",
		RootCAs:    p,
	})

	require.NoError(t, tconn.SetReadDeadline(time.Now().Add(time.Second)))
	require.Contains(t, tconn.Handshake().Error(), "timeout")
	sconn.Close()
}

func TestSconn_ExceedMTU(t *testing.T) {
	// echo server
	go func() {
		cert, err := tls.X509KeyPair([]byte(cert), []byte(key))
		require.NoError(t, err)

		conn, err := net.DialUDP("udp", &net.UDPAddr{Port: 19986}, &net.UDPAddr{Port: 19987})
		require.NoError(t, err)
		sconn := utls.NewSconn(conn)
		tconn := tls.Server(sconn, &tls.Config{ClientAuth: tls.NoClientCert, Certificates: []tls.Certificate{cert}})
		defer tconn.Close()

		require.NoError(t, tconn.SetReadDeadline(time.Now().Add(time.Second*2)))
		require.NoError(t, tconn.Handshake())
		require.NoError(t, tconn.SetReadDeadline(time.Time{}))

		var buf []byte = make([]byte, 2000)
		for {
			if n, err := tconn.Read(buf); err != nil {
				panic(err)
			} else {
				if n == 0 {
					t.Fatal(n)
				} else {
					m, err := tconn.Write(buf[:n])
					require.NoError(t, err)
					require.Equal(t, n, m)
				}
			}
		}
	}()

	time.Sleep(time.Second)

	// client
	conn, err := net.DialUDP("udp", &net.UDPAddr{Port: 19987}, &net.UDPAddr{Port: 19986})
	require.NoError(t, err)
	sconn := utls.NewSconn(conn)

	p := x509.NewCertPool()
	require.Equal(t, true, p.AppendCertsFromPEM([]byte(cert)))
	tconn := tls.Client(sconn, &tls.Config{
		InsecureSkipVerify: false,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
		},
		ServerName: "localhost",
		RootCAs:    p,
	})

	require.NoError(t, tconn.SetReadDeadline(time.Now().Add(time.Second)))
	require.NoError(t, tconn.Handshake())
	require.NoError(t, tconn.SetReadDeadline(time.Time{}))

	// hello world
	var sda = make([]byte, 65537)
	rand.Read(sda)
	n, err := tconn.Write(sda)
	require.NoError(t, err)
	require.Equal(t, 65537, n)

	var rda = make([]byte, 2000)
	for i := 0; i < 65537; {
		n, err = tconn.Read(rda)
		require.NoError(t, err)
		require.Equal(t, rda[:n], sda[i:i+n])
		i = i + n
	}
}
