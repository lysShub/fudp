package ioer

type Spin struct {
	ch chan struct{}
}

func NewSpin() *Spin {
	return &Spin{
		ch: make(chan struct{}),
	}
}

func (s *Spin) Wait() {
	<-s.ch
}

func (s *Spin) WaitChan() (ch *chan struct{}) {
	return &s.ch
}

func (s *Spin) Signal() {
	select {
	case s.ch <- struct{}{}:
	default:
		// nothing
	}
}
