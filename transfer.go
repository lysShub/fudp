package fudp

import (
	"fmt"
	"io"

	"github.com/lysShub/fudp/constant"
)

func (f *fudp) Write() {

}

func (f *fudp) write() error {
	var file *file
	if file = f.getFile(); file == nil {
		return io.EOF
	}

	// 同步
	var start int64
	file.add()

	//
	fmt.Println(start)

	return nil
}

func (f *fudp) writeSync(file *file) int64 {

	var da = make([]byte, constant.MTU, constant.MTU)
	// n := copy(da[0:], (*(*[8]byte)(unsafe.Pointer(&fs)))[:])

	// packet.Pack(da[:n:cap(da)],)
	fmt.Println(da)
	return 0
}

func (f *fudp) Reade() {}
