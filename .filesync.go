package fudp

// 文件同步不需要可靠, 因为它始终可以退化, 直至退化为start point为0
// 但是同步通知包中结束包必须是可靠的, 因为不是client必须接受到此数据包才能进入下一个step

import (
	"hash/crc64"
	"io"
	"unsafe"

	"github.com/lysShub/fudp/constant"
	"github.com/lysShub/fudp/packet"
)

const blockSize int64 = 128 << 10 // 256KB
const hashSize = 8                // 32位
var tab *crc64.Table = crc64.MakeTable(crc64.ECMA)

// SyncServer 文件同步server
func (f *fudp) SyncServer(file *file) (start uint64, err error) {

	var da = make([]byte, constant.MTU)
	var point int64 = 0 // 定位开始偏移指针

	var n int
	var tmp [4]byte
	var end bool = false
	var rbuf = make([]byte, blockSize)
	for !end {
		if n, err = f.rawConn.Read(da); err != nil {
			return 0, err
		} else {
			if ok, da := f.expSyncPackage(point, da[:n]); ok {
				l := len(da)
				end = da[l-1] == 1 // peer end
				copy(tmp[:], da[l-5:l-1])
				endBlockSize := int64(*(*int32)(unsafe.Pointer(&tmp)))
				endPoint := (int64(l-5)/hashSize)*blockSize + endBlockSize
				if endBlockSize > hashSize {
					// 异常
				} else if (l-5)%hashSize != 0 {
					// 异常
				}

				// 校验hash
				for index := 0; point < endPoint; point, index = point+blockSize, index+hashSize {
					if n, err = file.fh.WriteAt(rbuf, point); err == nil {
						crc64.Checksum(rbuf[:n], tab)

						if point+int64(n) == file.size { // 恰好读取完成时err为nil
							end = true
						}
					} else {

						if err == io.EOF {
							end = true
							if (endPoint - point) == int64(n) {

								// 如果校验成功
								point = point + int64(n)
							}
						} else {
							return 0, err // 文件读取出现错误, 对端只能超时退出
						}
					}
				}
			}
		}

		// 通知发送下一同步数据包
	}

	// 校验已经结束, 即已经确定start point

	return
}

func (f *fudp) SyncClient(file *file) {}

func (f *fudp) expSyncPackage(bias int64, da []byte) (bool, []byte) {
	da, bi, other, pt, err := packet.Parse(da, f.gcm)
	if (err == nil) && (bi == bias && other == 0 && pt == 1) && (len(da) > 5) { // 同步数据包大小大于5
		return true, da
	}
	return false, da
}

/*
  文件比较：
  Server一方拥有完整文件、是参照，其返回值为文件相同数据的最大偏移处
  Client是文件待校验的一方

  算法：
  256KB数据计算为16B的hash, 压缩比达到4096, 100GB的文件需要传输的数据。
*/
