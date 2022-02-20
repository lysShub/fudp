package packet_test

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/lysShub/fudp/packet"
)

type v struct {
	gcm cipher.AEAD

	length uint16
	fi     uint32
	bias   uint64
	pt     uint8
}

func TestMy(t *testing.T) {
	var data = []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 224}

	fmt.Println(packet.Parse(data, nil))
	t.Fatal(packet.Parse(data, nil))
}

func TestPacket(t *testing.T) {

	var tmp []byte = make([]byte, 16)
	rand.Read(tmp)
	var key [16]byte
	var none [12]byte
	copy(key[:], tmp)
	rand.Read(tmp)
	copy(none[:], tmp)
	gcm, err := Newgcm(key, none)
	if err != nil {
		t.Fatal(err)
	}

	var m []v = []v{
		{nil, 0, 0, 0, 0},
		{nil, 65506, uint32(1<<30 - 1), uint64(1<<64 - 1), uint8(1<<5 - 1)},
		{gcm, 0, 0, 0, 0},
		{gcm, 65506, uint32(1<<30 - 1), uint64(1<<64 - 1), uint8(1<<5 - 1)},
		randStruct(nil),
		randStruct(nil),
		randStruct(nil),
		randStruct(nil),
		randStruct(nil),
		randStruct(nil),
		randStruct(nil),
		randStruct(nil),
		randStruct(nil),
		randStruct(nil),
		randStruct(nil),
		randStruct(gcm),
		randStruct(gcm),
		randStruct(gcm),
		randStruct(gcm),
		randStruct(gcm),
		randStruct(gcm),
		randStruct(gcm),
		randStruct(gcm),
		randStruct(gcm),
	}

	for _, v := range m {
		test(t, &v)
	}

}

func BenchmarkPack(b *testing.B) {
	var length int = 1400

	var tmp []byte = make([]byte, 16)
	rand.Read(tmp)
	var key [16]byte
	var none [12]byte
	copy(key[:], tmp)
	rand.Read(tmp)
	copy(none[:], tmp)
	p, err := Newgcm(key, none)
	if err != nil {
		b.Fatal(err)
	}

	var tda []byte = make([]byte, length)
	rand.Read(tda)
	var da []byte = make([]byte, len(tda), len(tda)+29)

	var fi = uint32(randInt(0, 1<<30-1))
	var bias = uint64(randInt(0, 1<<63-1)) // 实际应该1<<64-1太大了容不下
	var pt = uint8(randInt(0, 1<<5-1))

	b.SetBytes(int64(length))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(da[:length], tda)
		packet.Pack(da[:length], fi, bias, pt, p)
	}
}

func BenchmarkParse(b *testing.B) {
	var length int = 1400

	var tmp []byte = make([]byte, 16)
	rand.Read(tmp)
	var key [16]byte
	var none [12]byte
	copy(key[:], tmp)
	rand.Read(tmp)
	copy(none[:], tmp)
	p, err := Newgcm(key, none)
	if err != nil {
		b.Fatal(err)
	}

	var tda []byte = make([]byte, length, length+29)
	rand.Read(tda)
	var fi = uint32(randInt(0, 1<<30-1))
	var bias = uint64(randInt(0, 1<<63-1)) // 实际应该1<<64-1太大了容不下
	var pt = uint8(randInt(0, 1<<5-1))
	ctl, err := packet.Pack(tda[:length], fi, bias, pt, p)
	if err != nil {
		b.Fatal(err)
	}

	var da []byte = make([]byte, len(tda))

	b.SetBytes(int64(length))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(da[:], tda[:ctl])
		packet.Parse(da, p)
	}
}

func BenchmarkEnGCM(b *testing.B) {
	var length int = 1400

	var tmp []byte = make([]byte, 16)
	rand.Read(tmp)
	var key [16]byte
	var none [12]byte
	copy(key[:], tmp)
	rand.Read(tmp)
	copy(none[:], tmp)

	var gcm cipher.AEAD
	if block, err := aes.NewCipher(key[:]); err != nil {
		b.Fatal(err)
	} else {
		if gcm, err = cipher.NewGCM(block); err != nil {
			b.Fatal(err)
		}
	}

	var tda []byte = make([]byte, length)
	rand.Read(tda)
	var da []byte = make([]byte, len(tda), len(tda)+29)
	var head []byte = make([]byte, 13)
	rand.Read(head)

	b.SetBytes(int64(length))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(da[:length], tda)
		da = gcm.Seal(da[:0], none[:], da[:length], head)
	}
}

func BenchmarkDeGCM(b *testing.B) {
	var length int = 1400

	var tmp []byte = make([]byte, 16)
	rand.Read(tmp)
	var key [16]byte
	var none [12]byte
	copy(key[:], tmp)
	rand.Read(tmp)
	copy(none[:], tmp)

	var gcm cipher.AEAD
	if block, err := aes.NewCipher(key[:]); err != nil {
		b.Fatal(err)
	} else {
		if gcm, err = cipher.NewGCM(block); err != nil {
			b.Fatal(err)
		}
	}

	var tda []byte = make([]byte, length, length+29)
	rand.Read(tda)
	var head []byte = make([]byte, 13)
	rand.Read(head)
	tda = gcm.Seal(tda[:0], none[:], tda[:length], head) // 密文
	var da []byte = make([]byte, len(tda))

	b.SetBytes(int64(length))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n := copy(da[:], tda)
		da = gcm.Seal(da[:0], none[:], da[:n], head)
	}
}

// 测试
func test(t *testing.T, v *v) {

	var tda []byte = make([]byte, v.length)
	rand.Read(tda)
	var da []byte = make([]byte, len(tda), cap(tda)+29)
	copy(da, tda)
	var n uint16
	var err error
	if n, err = packet.Pack(da, v.fi, v.bias, v.pt, v.gcm); err != nil {
		t.Fatal(err)
	}

	length1, fi1, bias1, pt1, err := packet.Parse(da[:n], v.gcm)
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

func randStruct(p cipher.AEAD) v {
	return v{
		gcm:    p,
		length: uint16(randInt(0, 65506)),
		fi:     uint32(randInt(0, 1<<30-1)),
		bias:   uint64(randInt(0, 1<<63-1)), // 实际应该1<<64-1太大了容不下
		pt:     uint8(randInt(0, 1<<5-1)),
	}
}

func Newgcm(key [16]byte, none [12]byte) (cipher.AEAD, error) {
	var gcm cipher.AEAD
	if block, err := aes.NewCipher(key[:]); err != nil {
		return nil, err
	} else {
		if gcm, err = cipher.NewGCM(block); err != nil {
			return nil, err
		}
	}
	return gcm, nil
}

func randInt(min, max int) int {
	b := new(big.Int).SetInt64(int64(max - min))
	i, _ := rand.Int(rand.Reader, b)
	return int(i.Int64()) + min
}
