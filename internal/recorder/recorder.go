package recorder

type Recorder struct {
	list []uint64 // 记录数据，奇数偶数组成一队，全闭

	gaps   uint64 // 所有间隙大小和
	blocks uint64 // 所有块个数

}

func NewRecorder() *Recorder {
	rs := &Recorder{
		list:   make([]uint64, 0, 128),
		gaps:   0,
		blocks: 0,
	}
	return rs
}

// [start end]
func (r *Recorder) Put(start, end uint64) {
	defer func() {
		if recover() != nil {
			r.list = append(r.list, start, end)
		}
	}()

	l := len(r.list)
	if start > end {
		return
	} else if start > r.list[l-1] { // diff大于0
		// 绝大多数情况
		if start-r.list[l-1] == 1 {
			r.list[l-1] = end
		} else {
			r.list = append(r.list, start, end)
		}
		return
	}

	var si, ei int = -1, -1 // block 索引位置
	for i := l - 1; i > 0; i = i - 2 {
		if r.list[i-1] <= end {

			// 新的的尾在此block有交集
			// 取两个尾的最大值
			if r.list[i] < end {
				r.list[i] = end // 吞并 swallow
			}
			ei = i
			if i+2 <= l-1 && r.list[i+1]-r.list[i] <= 1 {
				// 需要邻块合并
				n := copy(r.list[i:], r.list[i+2:])
				l, r.list = i+n, r.list[:i+n]
			}
			// 找到ei了

			// 开始找si
			for i = i + 0; i > 0; i = i - 2 {
				// if r.list[i-1] <= start { // 有大遍历到小,
				if r.list[i] < start { // 有大遍历到小,

					// 插入的头在此block有交集
					// 取两个头的最小值
					if start < r.list[i+1] {
						r.list[i+1] = start // 不可能执行到此处
					}
					si = i + 2
					if r.list[i+1]-r.list[i] <= 1 {
						// 需要邻块合并
						si, ei = si-2, ei-2
						r.list[i] = r.list[i+2]
						n := copy(r.list[i:], r.list[i+2:])
						l, r.list = i+n, r.list[:i+n]
					}

					goto mr // 可以合并
				}
			}

			// 除了第一个block，检查了所有的block; 没有找到
			si = 1
			if r.list[0] > start {
				r.list[0] = start
			}
			goto mr
		}
	}

mr:
	if si == -1 && ei == -1 {
		if end > r.list[l-1] {
			// 最后追加 不可能运行到
			r.list = append(r.list, start, end)
		} else {
			// 最前面
			if r.list[0]-end <= 1 {
				r.list[0] = start
			} else {
				r.list = append(r.list, 0, 0)
				copy(r.list[2:], r.list[0:])
				r.list[0], r.list[1] = start, end
			}
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
		s += r.list[i] - r.list[i-1] - 1
	}
	if len(r.list) > 0 {
		s += r.list[0] - 0
	}

	return s
}

// func (r *Recorder) Changed() int {

// }

func (r *Recorder) Show() []uint64 {
	return r.list
}
