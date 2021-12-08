package test

import (
	"testing"

	"github.com/lysShub/fudp"
	"github.com/magiconair/properties/assert"
)

func TestConfig(t *testing.T) {

	type Suite struct {
		fun    func(c *fudp.Config)
		exp    fudp.Config
		experr error // 部分测试用例期望错误
	}

	var suites = []Suite{
		{
			fun: func(c *fudp.Config) { c.CSMode() },
			exp: fudp.Config{
				// mode: 1,
			},
		},
	}

	for _, v := range suites {

		c, err := fudp.Configure(v.fun)
		if err != nil {
			if v.experr == nil {
				t.Fatal(err)
			} else {
				assert.Equal(t, err, v.experr)
			}
		} else {
			assert.Equal(t, c, v.exp)
		}

	}

}
