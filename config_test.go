package fudp_test

import (
	"testing"

	"github.com/lysShub/fudp"
)

func TestA(t *testing.T) {

	fudp.Configure(func(c *fudp.Config) {
		c.CSMode().Client()
		c.CSMode().Server([]byte("证书"), []byte("密钥"), nil)
		c.PPMode().Receive("/tmp")
		c.PPMode().Send("~/use/docs", true)
	})
}
