package recorder

type Recorder struct {
	list []uint64 // 记录数据，奇数偶数组成一队，全闭

	gaps   uint64 // 所有间隙大小和
	blocks uint64 // 所有块个数

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
	} else if start > r.list[len(r.list)-1] { // diff大于0
		// 绝大多数情况
		if start-r.list[len(r.list)-1] == 1 {
			r.list[len(r.list)-1] = end
		} else {
			r.list = append(r.list, start, end)
		}
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
			if i+2 <= len(r.list)-1 && r.list[i+1]-r.list[i] <= 1 {
				// 需要邻块合并
				n := copy(r.list[i:], r.list[i+2:])
				r.list = r.list[:i+n]
			}
			// 找到ei了

			// 开始找si
			for i = i + 0; i > 0; i = i - 2 {
				// if r.list[i] >= start {
				if r.list[i-1] <= start {

					// 插入的头在此block有交集
					// 取两个头的最小值
					if start <= r.list[i-1]-1 {
						r.list[i-1] = start
					}
					si = i
					if i-3 >= 0 && r.list[i-1]-r.list[i-2] <= 1 {
						// 需要邻块合并
						si, ei = si-2, ei-2
						n := copy(r.list[i-2:], r.list[i:])
						r.list = r.list[:i-2+n]
					}

					goto mr // 可以合并
				}
			}

			// 没找到si，不可能， 只能是start > end和
			// 已存在的最小值不为0，start 小于最小值
			// 两种情况
			si = 1
			r.list[0] = start
			goto mr
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
	s := uint64(0)
	for i := 2; i < len(r.list)-1; i = i + 2 {
		s += r.list[i] - r.list[i-1]
	}
	s += r.list[0] - 0
	return s
}

// func (r *Recorder) Changed() int {

// }

func (r *Recorder) Show() []uint64 {
	if r.list[1] == 1<<64-1 {
		return []uint64{}
	} else {
		return r.list
	}
}
