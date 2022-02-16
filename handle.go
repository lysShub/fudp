package fudp

// handle function
// 暂时仅实现完全匹配

import (
	"errors"
	"net/url"
	"sync"
)

// HandleFunc 注册全局handle
func HandleFunc(pattern string, handler Handler) error {
	return defaultServerMux.handleFunc(pattern, handler)
}

// Handle 获取handleFunc
func Handle(pattern string) Handler {
	if h, ok := defaultServerMux.m[pattern]; ok {
		return h
	}
	return nil
}

// path 磁盘路径
// path为Server机器上文件路径
type Handler func(url *url.URL) (path string, stateCode int)

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
