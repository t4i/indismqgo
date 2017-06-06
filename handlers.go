package indismqgo

import (
	"sync"
)

type Handlers struct {
	handlers map[string]Handler
	sync.RWMutex
}

func (hs *Handlers) Get(key string) (h Handler) {
	hs.RLock()
	h = hs.handlers[key]
	hs.RUnlock()
	return
}
func (hs *Handlers) Set(key string, val Handler) {
	hs.Lock()
	hs.handlers[key] = val
	hs.Unlock()

}
func (hs *Handlers) Del(key string) {
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
