package connections

import (
	"errors"
	"github.com/t4i/indismqgo"
	"sync"
)

var _ indismqgo.SenderStore = (*Connections)(nil)

type Connections struct {
	connections map[string]indismqgo.Sender
	lock        sync.RWMutex
}

func NewConnections() *Connections {
	c := &Connections{}
	c.connections = map[string]indismqgo.Sender{}
	return c
}

// func NewConnection(sender Sender, conn interface{}, connType uint8) *Connection {
// 	return &Connection{Sender: sender, Conn: conn}
// }

// func (cc *Connection) Send(m *MsgBuffer) error {
// 	return cc.Sender(m, cc)
// }

func (cs *Connections) Send(key string, m *indismqgo.MsgBuffer) error {
	cc := cs.GetSender(key)
	if cc == nil {
		return errors.New("No Connection Information Function")
	}
	err := cc.Send(m)
	if err != nil {
		return err
	}
	return nil
}

func (cs *Connections) GetSender(key string) indismqgo.Sender {
	cs.lock.RLock()
	val := cs.connections[key]
	cs.lock.RUnlock()
	return val
}
func (cs *Connections) SetSender(key string, val indismqgo.Sender) {
	cs.lock.Lock()
	cs.connections[key] = val
	cs.lock.Unlock()

}
func (cs *Connections) DelSender(key string) {
	if val := cs.GetSender(key); val != nil {
		cs.lock.Lock()
		delete(cs.connections, key)
		cs.lock.Unlock()

	}

}
func (cs *Connections) Length() (l int) {
	cs.lock.RLock()
	l = len(cs.connections)
	cs.lock.RUnlock()
	return
}
