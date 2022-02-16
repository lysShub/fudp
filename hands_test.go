package fudp

import (
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHandP2P(t *testing.T) {

	// 测试P2P握手
	// 发送端开放19986端口, 接收端开放19987端口
	rconn, err := net.DialUDP("udp", &net.UDPAddr{Port: 19987}, &net.UDPAddr{Port: 19986})
	require.Nil(t, err)
	defer rconn.Close()
	sconn, err := net.DialUDP("udp", &net.UDPAddr{Port: 19986}, &net.UDPAddr{Port: 19987})
	require.Nil(t, err)
	defer sconn.Close()

	u, err := url.Parse("fudp://localhost:19986/a/bb?mtu=1372")
	require.Nil(t, err)
	var si = &fudp{
		rawConn:  sconn,
		key:      [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		handleFn: func(url *url.URL) (path string, stateCode int) { return "", 200 },
	}
	var ri = &fudp{
		rawConn: rconn,
		key:     [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		url:     u,
	}

	// 发送方
	var e error
	go func() {
		e = si.handPong()
	}()
	statueCode, err := ri.handPing(nil)

	time.Sleep(time.Second)
	require.Nil(t, err)
	require.Nil(t, e)
	require.Equal(t, uint16(200), statueCode)
}
