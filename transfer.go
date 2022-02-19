package fudp

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/lysShub/fudp/constant"
)

func (f *fudp) Write() {
	var fch chan fi = make(chan fi, 2)
	go walk(f.wpath, fch)

	for fi := range fch {

		f.write(fi.fh, fi.name)

	}
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

func (f *fudp) Reade() {}
func (f *fudp) readeSync()

type fi struct {
	fh   *os.File
	name string // 分割符为"/"
}

// walk
func walk(rpath string, fch chan fi) (err error) {
	if rpath, err = filepath.Abs(rpath); err != nil {
		return err
	}
	filepath.Walk(rpath, func(path string, info fs.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			if fh, err := os.Open(path); err == nil {
				name, _ := filepath.Rel(rpath, path)
				fch <- fi{
					fh:   fh,
					name: name,
				}
				return nil
			}
		}

		// 不能正常读取的文件
		return nil
	})

	close(fch)
	return nil
}
