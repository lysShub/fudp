package packet_test

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"testing"

	"github.com/lysShub/fudp/packet"
	"github.com/stretchr/testify/require"
)

type suit struct {
	gcm         cipher.AEAD
	length      uint16 // 数据长度
	other       uint8
	packageType uint8
	err         error

	exp []byte
}

func TestMy(t *testing.T) {
	var key [16]byte
	rand.Read(key[:])
	var gcm cipher.AEAD
	block, err := aes.NewCipher(key[:])
	require.NoError(t, err)
	gcm, err = cipher.NewGCM(block)
	require.NoError(t, err)

	var da = make([]byte, 1500, 1500+packet.ExpendLen)
	da = packet.Pack(da, 1832, 9, 11, gcm)
	require.Equal(t, 1500+packet.ExpendLen, len(da))

	// []uint8 len: 25, cap: 32, [2,0,100,114,201,36,188,210,237,147,127,195,163,79,178,75,40,7,0,0,0,0,0,0,155]
}

func TestPacket(t *testing.T) {
	var key [16]byte
	rand.Read(key[:])
	var gcm cipher.AEAD
	block, err := aes.NewCipher(key[:])
	require.NoError(t, err)
	gcm, err = cipher.NewGCM(block)
	require.NoError(t, err)

	var da = make([]byte, 1500)
	rand.Read(da[:1500])
	da = packet.Pack(da, 1832, 9, 11, gcm)
	require.Equal(t, 1500+packet.ExpendLen, len(da))

	da, bias, other, pt, err := packet.Parse(da, gcm)
	require.NoError(t, err)
	require.Equal(t, 1500, len(da))
	require.Equal(t, 1832, int(bias))
	require.Equal(t, 9, int(other))
	require.Equal(t, 11, int(pt))
}
