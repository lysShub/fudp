package fudp_test

import (
	"net/url"
	"testing"

	"github.com/lysShub/fudp"
)

func TestA(t *testing.T) {
	var handleFunc = func(pars *url.URL) (path string, err error) {
		return "./", nil
	}

	fudp.Configure(func(c *fudp.Config) {
		c.CSMode().Client()
		c.CSMode().Server([]byte("证书"), []byte("密钥"), handleFunc)
		c.PPMode().Receive("/tmp")
		c.PPMode().Send("~/use/docs")
	})
}
