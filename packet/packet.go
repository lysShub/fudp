package packet

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"strconv"
	"sync"
)

type Packet struct {
	// 协议包格式/解析

	// lock    sync.Mutex
	sync.Mutex
	encrypt bool
	key     [16]byte
	none    [12]byte
	gcm     cipher.AEAD
}

// pack 打包, 确保data有足够的容量, 否则会打包失败
// 	@ data: 数据，其cap至少应该比len大29(13+16); 最大小不大于65506
// 	@ pt:	包类型
// 	@ bias:	数据偏置
// 	@ fi:	文件序号
// 返回包的有效长度
func (p *Packet) Pack(data []byte, fi uint32, bias uint64, pt uint8) (length uint16, err error) {

	if cap(data)-len(data) < 29 {
		return 0, errors.New("expect capacity of data more than length 29, actual len(data):" + strconv.Itoa(len(data)) + "   cap(data):" + strconv.Itoa(cap(data)))
	} else if fi > 0x3FFFFFFF {
		return 0, errors.New("invalid fi, expcet fi <=0x3FFFFFFF, actual " + strconv.FormatInt(int64(fi), 0xf))
	} else if pt > 0b11111 {
		return 0, errors.New("invalid pt, expcet pt <=0x1F, actual " + strconv.FormatInt(int64(pt), 16))
	} else if len(data) > 65506 { // UDP无法承受之重与uint16无法承受之重
		return 0, errors.New("expect length of parameter date not more than 65506, actual len(data):" + strconv.Itoa(len(data)))
	}

	var head []byte = make([]byte, 0, 13)

	var lfi, foo uint8 = 0, uint8(fi&0b111111) << 2
	fi = fi >> 6
	for i := 2; i >= 0; i-- {
		if fi>>(8*i) > 0 {
			head = append(head, uint8(fi>>(8*i)))
		} else {
			break
		}
	}
	lfi = uint8(len(head))
	head = append(head, foo+lfi&0b11)

	for i := 7; i >= 1; i-- {
		if bias>>(8*i) <= bias {
			head = append(head, uint8(bias>>(8*i)))
		} else {
			break
		}
	}
	head = append(head, uint8(bias))
	lbias := uint8(len(head)) - lfi - 1
	head = append(head, ((lbias-1)&0b111)<<5+pt&0b11111)

	hl := lfi + lbias + 2
	if p.encrypt {
		data = p.gcm.Seal(data[:0], p.none[:], data, head[:hl])
		data = append(data[:], head[:hl]...)
		return uint16(len(data)), nil
	} else {
		data = append(data, head[:hl]...)
		return uint16(len(data)), nil
	}
}

// parse 解包
func (p *Packet) Parse(data []byte) (length uint16, fi uint32, bias uint64, pt uint8, err error) {
	l := len(data) - 1

	if l >= 2 {
		pt = 0b11111 & data[l]
	} else {
		return 0, 0, 0, 0, errors.New("parse fail: package at least 3 Bytes")
	}

	var lbias, i uint8 = (data[l]&0b11100000)>>5 + 1, 0
	for l = l - 1; i < lbias && l > 0; l, i = l-1, i+1 {
		bias = bias + uint64(data[l])<<(8*i)
	}
	if lbias != i || l < 0 {
		return 0, 0, 0, 0, errors.New("parse fail: bias")
	}

	var lfi, j uint8 = data[l] & 0b11, 0
	fi = uint32(data[l]&0b11111100) >> 2
	for l = l - 1; j < lfi && l >= 0; l, j = l-1, j+1 {
		fi = fi + uint32(data[l])<<(j*8+6)
	}
	if j != lfi {
		return 0, 0, 0, 0, errors.New("parse fail: lfi")
	}

	if p.encrypt {
		data, err = p.gcm.Open(data[:0], p.none[:], data[:l+1], data[l+1:])
		if err != nil {
			length = 0
		} else {
			length = uint16(len(data))
		}
	} else {
		length = uint16(l) + 1
	}
	return
}

// SetKey 设置加密, 设置以后所有发送的包都会被加密、所有接收的包都会被解密
//	@key: 密钥  @none: 随机数       传输双方必须相同
func (p *Packet) SetKey(key [16]byte, none [12]byte) error {
	p.Lock()
	copy(p.key[:], key[:])
	copy(p.none[:], none[:])
	if block, err := aes.NewCipher(p.key[:]); err != nil {
		return err
	} else {
		if p.gcm, err = cipher.NewGCM(block); err != nil {
			return err
		}
	}
	p.encrypt = true
	p.Unlock()
	return nil
}
