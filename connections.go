package indismqgo

import (
	"errors"
	"sync"
)

type Connections struct {
	connections map[string]Connection
	lock        sync.RWMutex
}

// func NewConnection(sender Sender, conn interface{}, connType uint8) *Connection {
// 	return &Connection{Sender: sender, Conn: conn}
// }

// func (cc *Connection) Send(m *MsgBuffer) error {
// 	return cc.Sender(m, cc)
// }

func (cs *Connections) Send(key string, m *MsgBuffer) error {
	cc := cs.Get(key)
	if cc == nil {
		return errors.New("No Connection Information Function")
	}
	err := cc.Send(m)
	if err != nil {
		return err
	}
	return nil
}

func (cs *Connections) Get(key string) Connection {
	cs.lock.RLock()
	val := cs.connections[key]
	cs.lock.RUnlock()
	return val
}
func (cs *Connections) Set(key string, val Connection) {
	cs.lock.Lock()
	cs.connections[key] = val
	cs.lock.Unlock()

}
func (cs *Connections) Del(key string) {
	cs.lock.Lock()
	delete(cs.connections, key)
	cs.lock.Unlock()
}
func (cs *Connections) Length() (l int) {
	cs.lock.RLock()
	l = len(cs.connections)
	cs.lock.RUnlock()
	return
}
