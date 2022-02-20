package fudp

import "testing"

func TestPrePare(t *testing.T) {

	fu := &fudp{wpath: `./example`}

	fu.prepare()

}
