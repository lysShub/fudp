package main

import (
	"fmt"
	_ "net/http/pprof"

	"github.com/lysShub/fudp"
)

func main() {

	// fudp.New().PPMode().Send("")
	f := fudp.New(func(c *fudp.Configure) {
		c.PPMode().Receive("C:/lys", []byte("123"))
	})

	fmt.Println(*f)
	return

}
