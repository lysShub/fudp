package packet_test

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/lysShub/fudp/packet"
)

type v struct {
	p *packet.Packet

	length uint16
	fi     uint32
	bias   uint64
	pt     uint8
}

func TestPacket(t *testing.T) {

	var unEncrypt *packet.Packet = &packet.Packet{}
	var encrypt *packet.Packet = &packet.Packet{}

	var tmp []byte = make([]byte, 16)
	rand.Read(tmp)
	var key [16]byte
	var none [12]byte
	copy(key[:], tmp)
	rand.Read(tmp)
	copy(none[:], tmp)
	if err := encrypt.SetKey(key, none); err != nil {
		t.Fatal(err)
	}

	var m []v = []v{
		{unEncrypt, 0, 0, 0, 0},
		{unEncrypt, 65506, uint32(1<<30 - 1), uint64(1<<64 - 1), uint8(1<<5 - 1)},
		{encrypt, 0, 0, 0, 0},
		{encrypt, 65506, uint32(1<<30 - 1), uint64(1<<64 - 1), uint8(1<<5 - 1)},
		randStruct(unEncrypt),
		randStruct(unEncrypt),
		randStruct(unEncrypt),
		randStruct(unEncrypt),
		randStruct(unEncrypt),
		randStruct(unEncrypt),
		randStruct(unEncrypt),
		randStruct(unEncrypt),
		randStruct(unEncrypt),
		randStruct(unEncrypt),
		randStruct(unEncrypt),
		randStruct(encrypt),
		randStruct(encrypt),
		randStruct(encrypt),
		randStruct(encrypt),
		randStruct(encrypt),
		randStruct(encrypt),
		randStruct(encrypt),
		randStruct(encrypt),
		randStruct(encrypt),
	}

	for _, v := range m {
		test(t, &v)
	}

}

// 随机测试
func test(t *testing.T, v *v) {

	var tda []byte = make([]byte, v.length)
	rand.Read(tda)

	var da []byte = make([]byte, len(tda), cap(tda)+29)
	copy(da, tda)
	var n uint16
	var err error
	if n, err = v.p.Pack(da, v.fi, v.bias, v.pt); err != nil {
		t.Fatal(err)
	}

	length1, fi1, bias1, pt1, err := v.p.Parse(da[:n])
	if err != nil {
		t.Fatal(err)
	}
	if v.length != length1 {
		t.Fatal(fmt.Sprint("length ", v.length, length1))
	}
	if v.fi != fi1 {
		t.Fatal(fmt.Sprint("fi ", v.fi, fi1))
	}
	if v.bias != bias1 {
		t.Fatal(fmt.Sprint("bias ", v.bias, bias1))
	}
	if v.pt != pt1 {
		t.Fatal(fmt.Sprint("pt ", v.pt, pt1))
	}
	if !bytes.Equal(tda, da[:length1]) {
		t.Fatal("加密解密后不一样")
	}
}

func randStruct(p *packet.Packet) v {
	return v{
		p:      p,
		length: uint16(randInt(0, 65506)),
		fi:     uint32(randInt(0, 1<<30-1)),
		bias:   uint64(randInt(0, 1<<63-1)), // 实际应该1<<64-1太大了容不下
		pt:     uint8(randInt(0, 1<<5-1)),
	}
}

func randInt(min, max int) int {
	b := new(big.Int).SetInt64(int64(max - min))
	i, _ := rand.Int(rand.Reader, b)
	return int(i.Int64()) + min
}
