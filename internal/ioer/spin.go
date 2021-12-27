package ioer

// 容量必须为1
type spin chan struct{}

func (s spin) wait() {
	<-s
}

func (s spin) done() {
	select {
	case s <- struct{}{}:
	default:
	}
}
