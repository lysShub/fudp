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
	// require.Equal(t, firstSuit.gaps, gaps, "gaps: "+firstSuit.String())
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
	// base
	{
		put:    []uint64{0, 1372},
		exp:    []uint64{0, 1372},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{5, 1372},
		exp:    []uint64{5, 1372},
		blocks: 2,
		gaps:   4,
	},
	{
		put:    []uint64{1372, 0},
		exp:    []uint64{},
		blocks: 1,
		gaps:   0,
	},
	{
		put:    []uint64{1372, 5},
		exp:    []uint64{},
		blocks: 1,
		gaps:   0,
	},

	// 两个block
	{
		put:    []uint64{5, 1372, 0, 2},
		exp:    []uint64{0, 2, 5, 1372},
		blocks: 2,
		gaps:   3,
	},
	// {
	// 	put:    []uint64{5, 1372, 0, 4},
	// 	exp:    []uint64{0, 1372},
	// 	blocks: 1,
	// 	gaps:   1,
	// },
	{
		put:    []uint64{5, 1372, 0, 5},
		exp:    []uint64{0, 1372},
		blocks: 1,
		gaps:   0,
	},
}

var firstSuit suit = suit{
	// put:    []uint64{0, 1372, 1400, 2000, 2500, 3000, 6, 2600},
	// exp:    []uint64{0, 1372},

	put:    []uint64{5, 1372},
	exp:    []uint64{5, 1372},
	blocks: 1,
	gaps:   0,
}
