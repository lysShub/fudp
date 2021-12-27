package ioer_test

import (
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/lysShub/fudp/internal/ioer"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	var sAddr = &net.UDPAddr{IP: nil, Port: 1314}

	l, err := ioer.Listen("udp", sAddr)
	require.NoError(t, err)
	defer l.Close()
	go func() {
		for conn, err := l.Accept(); err == nil; {
			go func(c *ioer.Conn) {
				var da []byte = make([]byte, 128)
				for {
					n, err := c.Read(da)
					require.NoError(t, err)
					c.Write(da[:n]) // Pong
				}
			}(conn)
		}
	}()

	for i := 0; i < 48; i++ {
		go func() {
			da := make([]byte, 48)
			n, err := rand.Read(da)
			require.NoError(t, err)
			require.Equal(t, 48, n)
			conn, err := ioer.Dial("udp", nil, sAddr)
			require.NoError(t, err)
			defer conn.Close()
			n, err = conn.Write(da) // Ping
			require.NoError(t, err)
			require.Equal(t, 48, n)
			da2 := make([]byte, 48)
			n, err = conn.Read(da2)
			require.NoError(t, err)
			require.Equal(t, 48, n)

			require.Equal(t, da[:48], da2[:48])
		}()
	}
}

func TestDial(t *testing.T) {
	sda := make([]byte, 48)
	rda := make([]byte, 48)
	var sAddr = &net.UDPAddr{IP: nil, Port: 19986}

	l, err := ioer.Listen("udp", sAddr)
	require.NoError(t, err)
	defer l.Close()
	go func() {
		for conn, err := l.Accept(); err == nil; {
			go func(c *ioer.Conn) {
				var da []byte = make([]byte, 128)
				for {
					n, err := c.Read(da)
					require.NoError(t, err)
					time.Sleep(time.Millisecond * 1500)
					c.Write(da[:n]) // Pong
				}
			}(conn)
		}
	}()
	time.Sleep(time.Millisecond * 500)

	// normal
	conn, err := ioer.Dial("udp", nil, sAddr)
	require.NoError(t, err)
	_, err = rand.Read(sda)
	require.NoError(t, err)
	_, err = conn.Write(sda)
	require.NoError(t, err)
	err = conn.SetReadDeadline(time.Now().Add(time.Millisecond * 1600))
	require.NoError(t, err)
	start := time.Now()
	_, err = conn.Read(rda)
	require.NoError(t, err)
	require.Equal(t, 150, int(time.Since(start))/1e7) // 精度10ms
	require.Equal(t, sda, rda)
	require.NoError(t, conn.Close())

	// timeout
	conn, err = ioer.Dial("udp", nil, sAddr)
	require.NoError(t, err)
	_, err = rand.Read(sda)
	require.NoError(t, err)
	_, err = conn.Write(sda)
	require.NoError(t, err)
	err = conn.SetReadDeadline(time.Now().Add(time.Millisecond * 1400))
	require.NoError(t, err)
	start = time.Now()
	_, err = conn.Read(rda)
	require.Error(t, err)
	require.Contains(t, err.Error(), "timeout")
	require.Equal(t, 140, int(time.Since(start))/1e7) // 精度10ms
}
