package recorder_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lysShub/fudp/internal/recorder"

	"github.com/stretchr/testify/require"
)

func TestRecorder(t *testing.T) {
	var list []uint64
	// var blocks uint64
	// var gaps uint64
	r := recorder.NewRecorder()
	for i := 0; i < len(firstSuit.put)-1; i = i + 2 {
		r.Put(firstSuit.put[i], firstSuit.put[i+1])
		list = r.Show()
		list = r.Show()
		// blocks = r.Blocks()
		// gaps = r.GapSize()
		// gaps = r.GapSize()

	}

	require.Equal(t, firstSuit.exp, list, "list: "+firstSuit.String())
	// require.Equal(t, firstSuit.blocks, blocks, "blocks: "+firstSuit.String())
	// require.Equal(t, firstSuit.gaps, r.GapSize(), "gaps: "+firstSuit.String())
	//

	//
	for _, suit := range data {
		r := recorder.NewRecorder()
		for i := 0; i < len(suit.put)-1; i = i + 2 {
			r.Put(suit.put[i], suit.put[i+1])
		}
		require.Equal(t, suit.exp, r.Show(), "list: "+suit.String())
		// require.Equal(t, suit.blocks, r.Blocks(), "blocks: "+suit.String())
		require.Equal(t, suit.gaps, r.GapSize(), "gaps: "+suit.String())
	}
}

type suit struct {
	put []uint64 // 奇偶组成一个block
	exp []uint64

	blocks uint64
	gaps   uint64
}

func (s suit) String() string {
	r := strings.ReplaceAll(fmt.Sprint(s.put), "]", "}")
	r = strings.ReplaceAll(r, "[", "{")
	r = strings.ReplaceAll(r, " ", ", ")
	return r
}

var data []suit = []suit{
	// new
	{
		put:    []uint64{0, 9},
		exp:    []uint64{0, 9},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{9, 0},
		exp:    []uint64{},
		blocks: 0,
		gaps:   0,
	},
	{
		put:    []uint64{1, 9},
		exp:    []uint64{1, 9},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{9, 1},
		exp:    []uint64{},
		blocks: 0,
		gaps:   0,
	},

	// 第二个block
	{
		put:    []uint64{0, 9, 0, 0},
		exp:    []uint64{0, 9},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 0, 1},
		exp:    []uint64{0, 9},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 0, 5},
		exp:    []uint64{0, 9},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 0, 10},
		exp:    []uint64{0, 10},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 0, 15},
		exp:    []uint64{0, 15},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 1, 1},
		exp:    []uint64{0, 9},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 1, 5},
		exp:    []uint64{0, 9},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 1, 8},
		exp:    []uint64{0, 9},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 1, 9},
		exp:    []uint64{0, 9},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 1, 10},
		exp:    []uint64{0, 10},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 1, 15},
		exp:    []uint64{0, 15},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 5, 5},
		exp:    []uint64{0, 9},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 5, 8},
		exp:    []uint64{0, 9},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 5, 9},
		exp:    []uint64{0, 9},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 5, 10},
		exp:    []uint64{0, 10},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 5, 15},
		exp:    []uint64{0, 15},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 8, 8},
		exp:    []uint64{0, 9},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 8, 9},
		exp:    []uint64{0, 9},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 8, 10},
		exp:    []uint64{0, 10},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 8, 15},
		exp:    []uint64{0, 15},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 9, 9},
		exp:    []uint64{0, 9},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 9, 10},
		exp:    []uint64{0, 10},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 9, 15},
		exp:    []uint64{0, 15},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 10, 10},
		exp:    []uint64{0, 10},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 10, 15},
		exp:    []uint64{0, 15},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{0, 9, 15, 15},
		exp:    []uint64{0, 9, 15, 15},
		blocks: 2,
		gaps:   5,
	},
	{
		put:    []uint64{0, 9, 15, 20},
		exp:    []uint64{0, 9, 15, 20},
		blocks: 2,
		gaps:   5,
	},
	//
	{
		put:    []uint64{1, 9, 0, 0},
		exp:    []uint64{0, 9},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{1, 9, 0, 1},
		exp:    []uint64{0, 9},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{1, 9, 0, 5},
		exp:    []uint64{0, 9},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{1, 9, 0, 10},
		exp:    []uint64{0, 10},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{1, 9, 0, 15},
		exp:    []uint64{0, 15},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{1, 9, 1, 1},
		exp:    []uint64{1, 9},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 1, 5},
		exp:    []uint64{1, 9},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 1, 8},
		exp:    []uint64{1, 9},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 1, 9},
		exp:    []uint64{1, 9},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 1, 10},
		exp:    []uint64{1, 10},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 1, 15},
		exp:    []uint64{1, 15},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 5, 5},
		exp:    []uint64{1, 9},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 5, 8},
		exp:    []uint64{1, 9},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 5, 9},
		exp:    []uint64{1, 9},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 5, 10},
		exp:    []uint64{1, 10},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 5, 15},
		exp:    []uint64{1, 15},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 8, 8},
		exp:    []uint64{1, 9},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 8, 9},
		exp:    []uint64{1, 9},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 8, 10},
		exp:    []uint64{1, 10},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 8, 15},
		exp:    []uint64{1, 15},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 9, 9},
		exp:    []uint64{1, 9},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 9, 10},
		exp:    []uint64{1, 10},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 9, 15},
		exp:    []uint64{1, 15},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 10, 10},
		exp:    []uint64{1, 10},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 10, 15},
		exp:    []uint64{1, 15},
		blocks: 1,
		gaps:   1,
	},
	{
		put:    []uint64{1, 9, 15, 15},
		exp:    []uint64{1, 9, 15, 15},
		blocks: 2,
		gaps:   6,
	},
	{
		put:    []uint64{1, 9, 15, 20},
		exp:    []uint64{1, 9, 15, 20},
		blocks: 2,
		gaps:   6,
	},

	// 三个block 跨block
	{
		put:    []uint64{5, 9, 15, 20, 0, 4},
		exp:    []uint64{0, 9, 15, 20},
		blocks: 2,
		gaps:   5,
	},
	{
		put:    []uint64{5, 9, 15, 20, 0, 4},
		exp:    []uint64{0, 9, 15, 20},
		blocks: 2,
		gaps:   5,
	},
	{
		put:    []uint64{5, 9, 15, 20, 0, 5},
		exp:    []uint64{0, 9, 15, 20},
		blocks: 2,
		gaps:   5,
	},
	{
		put:    []uint64{5, 9, 15, 20, 0, 6},
		exp:    []uint64{0, 9, 15, 20},
		blocks: 2,
		gaps:   5,
	},
	{
		put:    []uint64{5, 9, 15, 20, 0, 9},
		exp:    []uint64{0, 9, 15, 20},
		blocks: 2,
		gaps:   5,
	},
	{
		put:    []uint64{5, 9, 15, 20, 0, 10},
		exp:    []uint64{0, 10, 15, 20},
		blocks: 2,
		gaps:   4,
	},
	{
		put:    []uint64{5, 9, 15, 20, 0, 12},
		exp:    []uint64{0, 12, 15, 20},
		blocks: 2,
		gaps:   2,
	},
	{
		put:    []uint64{5, 9, 15, 20, 0, 12},
		exp:    []uint64{0, 12, 15, 20},
		blocks: 2,
		gaps:   2,
	},
	{
		put:    []uint64{5, 9, 15, 20, 0, 14},
		exp:    []uint64{0, 20},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{5, 9, 15, 20, 0, 15},
		exp:    []uint64{0, 20},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{5, 9, 15, 20, 0, 16},
		exp:    []uint64{0, 20},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{5, 9, 15, 20, 0, 19},
		exp:    []uint64{0, 20},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{5, 9, 15, 20, 0, 20},
		exp:    []uint64{0, 20},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{5, 9, 15, 20, 0, 21},
		exp:    []uint64{0, 21},
		blocks: 1,
		gaps:   0,
	},

	// 特殊情况
	{
		put:    []uint64{5, 9, 11, 15, 17, 18, 10, 16},
		exp:    []uint64{5, 18},
		blocks: 1,
		gaps:   5,
	},
	{
		put:    []uint64{1, 9, 11, 15, 17, 18, 10, 16, 0, 0},
		exp:    []uint64{0, 18},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{1, 5, 8, 11, 15, 18, 13, 16},
		exp:    []uint64{1, 5, 8, 11, 13, 18},
		blocks: 3,
		gaps:   4,
	},
}

var firstSuit suit = suit{
	put:    []uint64{5, 9, 15, 20, 0, 15},
	exp:    []uint64{0, 20},
	blocks: 1,
	gaps:   0,
}
