package fudp

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/lysShub/fudp/constant"
)

// 支持并行传输，当前仅支持均速并行
// 尝试发送方使用一个全局的write
//
// windwos 并行数为1

func (f *fudp) Write() {

}

type write struct {
	speedfactor int // 传输系数
}

func (f *fudp) write(fh *os.File, name string) {

	fstat, _ := fh.Stat()
	// sync
	fmt.Println(fstat)

}

func (f *fudp) writeSync(fs uint64) {
	// 同步的pt=0, fi递增； 0值fi发送得文件大小
	var da = make([]byte, constant.MTU, mcap)
	n := copy(da[0:], (*(*[8]byte)(unsafe.Pointer(&fs)))[:])

	// packet.Pack(da[:n:cap(da)],)
	fmt.Println(n)

}

func (f *fudp) Reade()     {}
func (f *fudp) readeSync() {}
