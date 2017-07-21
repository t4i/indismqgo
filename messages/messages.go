package messages

import (
	"github.com/t4i/indismqgo"
	"sync"
)

var _ indismqgo.MessageStore = (*Messages)(nil)

type Messages struct {
	messages      map[string]*indismqgo.MsgBuffer
	storeTime     map[int64]map[*indismqgo.MsgBuffer]bool
	UseStoreTime  bool
	storeTimeLock sync.RWMutex
	sync.RWMutex
}

func NewMessages() *Messages {
	m := &Messages{}
	m.messages = map[string]*indismqgo.MsgBuffer{}
	m.storeTime = map[int64]map[*indismqgo.MsgBuffer]bool{}
	return m
}
func (ms *Messages) GetMessage(key string) (m *indismqgo.MsgBuffer) {
	ms.RLock()
	m = ms.messages[key]
	ms.RUnlock()
	return
}

func (ms *Messages) SetMessageById(m *indismqgo.MsgBuffer) {
	ms.SetMessage(string(m.Fields.MsgId()), m)

}

func (ms *Messages) SetMessage(key string, m *indismqgo.MsgBuffer) {
	if m.Callback != nil {
		if ms.UseStoreTime && m.Timestamp > 0 {
			ms.storeTimeLock.Lock()
			if v, ok := ms.storeTime[m.Timestamp]; ok {
				v[m] = true
			} else {
				ms.storeTime[m.Timestamp] = make(map[*indismqgo.MsgBuffer]bool)
			}
			ms.storeTimeLock.Unlock()
		}
		ms.Lock()
		ms.messages[key] = m
		ms.Unlock()
	}

}
func (ms *Messages) DelMessage(key string) {
	if ms.UseStoreTime {
		if m := ms.GetMessage(key); m != nil && m.Timestamp > 0 {
			ms.storeTimeLock.Lock()
			if v, ok := ms.storeTime[m.Timestamp]; ok {
				if _, ok2 := v[m]; ok2 {
					if len(v) < 2 {
						delete(ms.storeTime, m.Timestamp)
					} else {
						delete(v, m)
					}
				}
			} else {
				ms.storeTime[m.Timestamp] = make(map[*indismqgo.MsgBuffer]bool)
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
