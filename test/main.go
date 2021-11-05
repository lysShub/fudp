package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	_ "net/http/pprof"

	"github.com/lysShub/fudp/packet"
)

func main() {
	var encrypt *packet.Packet = &packet.Packet{}
	var tmp []byte = make([]byte, 16)
	rand.Read(tmp)
	var key [16]byte
	var none [12]byte
	copy(key[:], tmp)
	rand.Read(tmp)
	copy(none[:], tmp)
	if err := encrypt.SetKey(key, none); err != nil {
		panic(err)
	}

	test(&v{encrypt, 65506, uint32(1<<30 - 1), uint64(1<<64 - 1), uint8(1<<5 - 1)})
}

type v struct {
	p *packet.Packet

	length uint16
	fi     uint32
	bias   uint64
	pt     uint8
}

func test(v *v) {

	var tda []byte = make([]byte, v.length)
	// rand.Read(tda)

	var da []byte = make([]byte, len(tda), cap(tda)+29)
	copy(da, tda)
	var n uint16
	var err error
	if n, err = v.p.Pack(da, v.fi, v.bias, v.pt); err != nil {
		panic(err)
	}

	length1, fi1, bias1, pt1, err := v.p.Parse(da[:n])
	if err != nil {
		panic(err)
	}
	if v.length != length1 {
		panic(fmt.Sprint("length ", v.length, length1))
	}
	if v.fi != fi1 {
		panic(fmt.Sprint("fi ", v.fi, fi1))
	}
	if v.bias != bias1 {
		panic(fmt.Sprint("bias ", v.bias, bias1))
	}
	if v.pt != pt1 {
		panic(fmt.Sprint("pt ", v.pt, pt1))
	}
	if !bytes.Equal(tda, da[:length1]) {
		panic("加密解密后不一样")
	}
}
