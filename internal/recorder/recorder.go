package recorder

type Recorder struct {
	list []uint64 // 记录数据，奇数偶数组成一队，全闭

	gaps   uint64 // 所有间隙大小和
	blocks uint64 // 所有块个数

	same bool
}

func NewRecorder() *Recorder {
	rs := &Recorder{
		list:   make([]uint64, 2, 128),
		gaps:   0,
		blocks: 0,
	}
	rs.list[0], rs.list[1] = 0, 1<<64-1
	return rs
}

// [start end]
func (r *Recorder) Put(start, end uint64) {
	if start > end {
		return
	}

	var si, ei int = -1, -1 // block 索引位置
	for i := len(r.list) - 1; i > 0; i = i - 2 {
		if r.list[i-1] <= end {

			// 新的的尾在此block有交集
			// 取两个尾的最大值
			if r.list[i]+1 <= end {
				r.list[i] = end // 吞并 swallow
			}
			ei = i
			// 找到ei了

			// 开始找si
			for i = i + 0; i > 0; i = i - 2 {
				if r.list[i] >= start {
					// 插入的头在此block有交集
					// 取两个头的最小值
					if start <= r.list[i-1]-1 {
						r.list[i-1] = start
					}
					si = i

					goto mr // 可以合并
				}
			}

			// 没找到si，不可能， 只能是start > end的情况
			return
		}
	}

mr:
	if si == -1 && ei == -1 {
		if end > r.list[len(r.list)-1] {
			// 最后追加
			r.list = append(r.list, start, end)
		} else {
			// 最前面
			r.list = append(r.list, 0, 0)
			copy(r.list[2:], r.list[0:])
			r.list[0], r.list[1] = start, end
		}
	} else if si == ei {
		return
	} else {
		// 合并

		n := copy(r.list[si:], r.list[ei:])
		r.list = r.list[:si+n]
	}

}

func (r *Recorder) Blocks() uint64 {
	return r.blocks
}
func (r *Recorder) GapSize() uint64 {
	return r.gaps
}

func (r *Recorder) Show() []uint64 {
	if r.list[1] == 1<<64-1 {
		return []uint64{}
	} else {
		return r.list
	}
}
