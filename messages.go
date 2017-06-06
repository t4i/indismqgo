package indismqgo

import (
	"sync"
)

type Messages struct {
	messages      map[string]*MsgBuffer
	storeTime     map[int64]map[*MsgBuffer]bool
	UseStoreTime  bool
	storeTimeLock sync.RWMutex
	sync.RWMutex
}

func (ms *Messages) Get(key string) (m *MsgBuffer) {
	ms.RLock()
	m = ms.messages[key]
	ms.RUnlock()
	return
}

func (ms *Messages) Store(m *MsgBuffer) {
	ms.Set(string(m.Fields.MsgId()), m)

}

func (ms *Messages) Set(key string, m *MsgBuffer) {
	if m.Callback != nil {
		if ms.UseStoreTime && m.timestamp > 0 {
			ms.storeTimeLock.Lock()
			if v, ok := ms.storeTime[m.timestamp]; ok {
				v[m] = true
			} else {
				ms.storeTime[m.timestamp] = make(map[*MsgBuffer]bool)
			}
			ms.storeTimeLock.Unlock()
		}
		ms.Lock()
		ms.messages[key] = m
		ms.Unlock()
	}

}
func (ms *Messages) Del(key string) {
	if ms.UseStoreTime {
		if m := ms.Get(key); m != nil && m.timestamp > 0 {
			ms.storeTimeLock.Lock()
			if v, ok := ms.storeTime[m.timestamp]; ok {
				if _, ok2 := v[m]; ok2 {
					if len(v) < 2 {
						delete(ms.storeTime, m.timestamp)
					} else {
						delete(v, m)
					}
				}
			} else {
				ms.storeTime[m.timestamp] = make(map[*MsgBuffer]bool)
			}
			ms.storeTimeLock.Unlock()
		}
	}
	ms.Lock()
	delete(ms.messages, key)
	ms.Unlock()
}
func (ms *Messages) Length() (l int) {
	ms.RLock()
	l = len(ms.messages)
	ms.RUnlock()
	return
}
