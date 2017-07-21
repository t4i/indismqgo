package handlers

import (
	"github.com/t4i/indismqgo"
	"sync"
)

var _ indismqgo.HandlerStore = (*Handlers)(nil)

type Handlers struct {
	handlers map[string]indismqgo.Handler
	sync.RWMutex
}

func NewHandlers() *Handlers {
	h := &Handlers{}
	h.handlers = map[string]indismqgo.Handler{}
	return h
}

func (hs *Handlers) GetHandler(key string) (h indismqgo.Handler) {
	hs.RLock()
	h = hs.handlers[key]
	hs.RUnlock()
	return
}
func (hs *Handlers) SetHandler(key string, val indismqgo.Handler) {
	hs.Lock()
	hs.handlers[key] = val
	hs.Unlock()

}
func (hs *Handlers) DelHandler(key string) {
	hs.Lock()
	delete(hs.handlers, key)
	hs.Unlock()
}
func (hs *Handlers) Length() (l int) {
	hs.RLock()
	l = len(hs.handlers)
	hs.RUnlock()
	return
}
