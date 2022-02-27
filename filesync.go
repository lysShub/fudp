package fudp

import (
	"strings"
	"time"
	"unsafe"

	"github.com/lysShub/fudp/constant"
	"github.com/lysShub/fudp/packet"
)

const filesyncPt = 3
const step = 64 << 20

// rewriteName 解决跨系统传输时，文件/文件夹命名冲突
var rewriteName func(path string) string = func(path string) string { return path }

// SyncServer 文件同步server
func (f *fudp) SyncServer(file *file) (start uint64, err error) {
	var da []byte = make([]byte, constant.MTU)
	var tmp1 [8]byte = *(*[8]byte)(unsafe.Pointer(&file.size))
	var tmp2 [4]byte = *(*[4]byte)(unsafe.Pointer(&file.mode))

	n := copy(da[0:], tmp1[:])
	n += copy(da[n:], tmp2[:])
	n += copy(da[n:], []byte(file.name))

	// 开始同步
	var point int = 0 // point以前的数据已经同步完成
	var syncFileinfoPack bool = false
	for {
		if err = f.rawConn.SetDeadline(time.Now().Add(constant.RTT)); err != nil {
			return 0, err
		}
		if _, err = f.rawConn.Write(packet.Pack(da[n:cap(da)], 0, 0, filesyncPt, f.gcm)); err != nil {
			return 0, err
		}

		if n, err = f.rawConn.Read(da); err != nil {
			if strings.Contains(err.Error(), "timeout") && !syncFileinfoPack { // 注意会陷入死循环
				time.Sleep(constant.RTT >> 1)
				continue
			}
			return 0, err
		} else {
			if ok, da := f.expSyncPackage(1, da[:n]); ok {
				syncFileinfoPack = true

				if len(da) > 0 {
					ptr := unsafe.Pointer(&da)
					for j := 0; j < len(da); j = j + 24 {
						start, end := *(*int64)(unsafe.Add(ptr, 0)), *(*int64)(unsafe.Add(ptr, 8))
						hash := *(*uint64)(unsafe.Add(ptr, 16))

						for x := start; x < end; x++ {

						}

					}
				} else {
					// peer end
				}

			}
		}

	}

}

// SyncServer 文件同步server
func (f *fudp) syncClient(file *file) {}

// expSyncPackage
func (f *fudp) expSyncPackage(index int64, da []byte) (bool, []byte) {
	da, bi, other, pt, err := packet.Parse(da, f.gcm)
	if (err == nil) && (bi == index && other == 0 && pt == filesyncPt) {
		return true, da
	}
	return false, da
}
