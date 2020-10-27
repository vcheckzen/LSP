package main

import "sync"

type syncMap struct {
	m map[interface{}]interface{}
	sync.RWMutex
}

func newSyncMap() *syncMap {
	return &syncMap{m: make(map[interface{}]interface{})}
}

func (s *syncMap) add(k, v interface{}) {
	s.Lock()
	s.m[k] = v
	s.Unlock()
}

func (s *syncMap) remove(k interface{}) {
	s.Lock()
	delete(s.m, k)
	s.Unlock()
}

func (s *syncMap) get(k interface{}) (v interface{}, ok bool) {
	s.RLock()
	v, ok = s.m[k]
	s.RUnlock()
	return
}

func (s *syncMap) clear() {
	s.Lock()
	s.m = make(map[interface{}]interface{})
	s.Unlock()
}

func (s *syncMap) isEmpty() bool {
	s.RLock()
	l := len(s.m)
	s.RUnlock()
	return l == 0
}

func (s *syncMap) keys() []interface{} {
	s.RLock()
	ks := make([]interface{}, 0, len(s.m))
	for k := range s.m {
		ks = append(ks, k)
	}
	s.RUnlock()
	return ks
}

func (s *syncMap) values() []interface{} {
	s.RLock()
	vs := make([]interface{}, 0, len(s.m))
	for _, v := range s.m {
		vs = append(vs, v)
	}
	s.RUnlock()
	return vs
}
