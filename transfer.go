package fudp

import (
	"fmt"
	"io"

	"github.com/lysShub/fudp/constant"
)

// 支持并行传输，当前仅支持均速并行
// 尝试发送方使用一个全局的write
//
// windwos 并行数为1

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
	start = writeSync(file)

	//
	fmt.Println(start)

	return nil
}

func writeSync(file *file) int64 {
	// 同步的pt=0, fi递增； 0值fi发送得文件大小
	var da = make([]byte, constant.MTU, mcap)
	// n := copy(da[0:], (*(*[8]byte)(unsafe.Pointer(&fs)))[:])

	// packet.Pack(da[:n:cap(da)],)
	fmt.Println(da)
	return 0
}

func (f *fudp) Reade()     {}
func (f *fudp) readeSync() {}
