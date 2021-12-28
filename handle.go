package fudp

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

// path 请求对应的工作路径
// stateCode 状态码, 参考HTTP
// msg 回复信息, 此信息不会加密
// 当path不为空时表示接受请求, 将继续通信
type Handler func(url *url.URL) (path string, stateCode int, msg string)

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
