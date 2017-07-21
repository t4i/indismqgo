package subscribers

import (
	"github.com/t4i/indismqgo"
	"sync"
)

var _ indismqgo.SubscriberStore = (*Subscribers)(nil)

type Subscribers struct {
	subscribers map[string]map[string]bool
	sync.RWMutex
}

func NewSubscribers() *Subscribers {
	return &Subscribers{subscribers: make(map[string]map[string]bool)}
}

func (ss *Subscribers) GetSubscriberList(key string) map[string]bool {
	ss.RLock()
	m := ss.subscribers[key]
	ss.RUnlock()
	return m
}
func (ss *Subscribers) SetSubscriber(key string, val map[string]bool) {
	ss.Lock()
	ss.subscribers[key] = val
	ss.Unlock()
}
func (ss *Subscribers) AddChannel(key string) {
	ss.Lock()
	ss.subscribers[key] = map[string]bool{}
	ss.Unlock()
}

func (ss *Subscribers) DelChannel(key string) {
	ss.Lock()
	delete(ss.subscribers, key)
	ss.Unlock()
}
func (ss *Subscribers) Length() (l int) {
	ss.RLock()
	l = len(ss.subscribers)
	ss.RUnlock()
	return
}

func (ss *Subscribers) AddSubscriber(client string, path string) {
	var present bool
	ss.Lock()
	if _, present = ss.subscribers[path]; !present {
		ss.subscribers[path] = make(map[string]bool)
	}
	ss.subscribers[path][client] = false
	ss.Unlock()
}

//DelSubscriber ...
func (ss *Subscribers) DelSubscriber(client string, path string) {
	var present, present2 bool
	ss.Lock()
	if _, present = ss.subscribers[path]; present {
		if _, present2 = ss.subscribers[path][client]; present2 {
			delete(ss.subscribers[path], client)
		}
		if len(ss.subscribers[path]) < 1 {
			delete(ss.subscribers, path)
		}
	}

	ss.Unlock()
}

//DelSubscriberAll ...
func (ss *Subscribers) DelSubscriberAll(client string) {
	for k := range ss.subscribers {
		ss.DelSubscriber(client, k)
	}
}
