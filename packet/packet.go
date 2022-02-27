package packet

import (
	"crypto/cipher"
	"errors"
	"unsafe"

	"github.com/lysShub/fudp/constant"
)

var none []byte = make([]byte, 12)

// Pack
// data: 数据本身, 确保cap(da)-len(da) > 9+16, 否则会重新分配内存
// bias: 偏移
// other: 备用字段, 取低4位
// packageType: 数据包类型, 取低4位
func Pack(data []byte, bias uint64, other uint8, packageType uint8, gcm cipher.AEAD) (packet []byte) {
	var head [9]byte
	head[8] = ((other & 0b1111) << 4) + packageType&0b1111
	copy(head[0:], (*(*[8]byte)(unsafe.Pointer(&bias)))[:])

	if gcm != nil {
		data = gcm.Seal(data[:0], none, data, head[:])
	}
	data = append(data, head[0:]...)
	return data
}

var ErrPacketFormat = errors.New("invalid package format")

func Parse(packet []byte, gcm cipher.AEAD) (data []byte, bias int64, other uint8, packageType uint8, err error) {
	l := len(packet)
	if l < constant.HeadSize {
		return nil, 0, 0, 0, ErrPacketFormat
	}

	other, packageType = packet[l-1]>>4, packet[l-1]&0b1111
	tmp := packet[l-9 : l-1]
	bias = *(*int64)(*(*unsafe.Pointer)(unsafe.Pointer(&tmp)))

	if gcm != nil {
		data, err = gcm.Open(packet[:0], none, packet[:l-constant.HeadSize], packet[l-constant.HeadSize:])
		if err != nil {
			if l-constant.HeadSize < 16 {
				return nil, 0, 0, 0, ErrPacketFormat
			}
			return nil, 0, 0, 0, err
		}
	} else {
		data = packet[:l-constant.HeadSize]
	}

	return
}
