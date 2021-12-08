package fudp

import (
	"testing"

	"github.com/lysShub/fudp/utils"
	"github.com/magiconair/properties/assert"
)

func TestConfig(t *testing.T) {

	type Suite struct {
		fun    func(c *Config)
		exp    Config
		experr error // 部分测试用例期望错误
	}

	var suites = []Suite{
		{
			fun: func(c *Config) { c.CSMode() },
			exp: Config{
				mode: CSMode,
			},
		},
		{
			fun: func(c *Config) { c.CSMode().Client() },
			exp: Config{
				mode: CSMode,
				role: CRole,
			},
		},
		{
			fun: func(c *Config) { c.CSMode().Client([]byte("root certificate")) },
			exp: Config{
				mode:     CSMode,
				role:     CRole,
				selfCert: [][]byte{[]byte("root certificate")},
			},
		},
		{
			fun: func(c *Config) { c.CSMode().Client([]byte("root certificate")).Receive("./") },
			exp: Config{
				mode:        CSMode,
				role:        CRole,
				acti:        DownloadAct,
				receivePath: utils.FormatPath(""),
				selfCert:    [][]byte{[]byte("root certificate")},
			},
		},
	}

	for _, v := range suites {

		c, err := Configure(v.fun)
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
