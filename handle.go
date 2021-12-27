package fudp

import (
	"errors"
	"net/url"
	"sync"
)

func HandleFunc(pattern string, handler Handler) error {
	return defaultServerMux.handleFunc(pattern, handler)
}

type Handler func(url *url.URL) (path string, err error)

// 现在只实现了完全匹配
type ServeMux struct {
	mu sync.RWMutex
	m  map[string]Handler
	es []string // 根据uri升序排序, 最长匹配
}

var defaultServerMux = &ServeMux{
	mu: sync.RWMutex{},
	m:  map[string]Handler{},
}

func (s *ServeMux) handleFunc(pattern string, handler Handler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(pattern) == 0 {
		return errors.New("fudp: invalid pattern")
	}
	if handler == nil {
		return errors.New("fudp: invalid handler function")
	}

	if _, exist := s.m[pattern]; exist {
		return errors.New("fudp: '" + pattern + "' registered")
	} else {
		s.m[pattern] = handler
	}
	return nil
}
