package connections

import (
	"errors"
	"github.com/t4i/indismqgo"
	"sync"
)

var _ indismqgo.ConnectionStore = (*Connections)(nil)

type Connections struct {
	connections map[string]indismqgo.Connection
	lock        sync.RWMutex
}

func NewConnections() *Connections {
	c := &Connections{}
	c.connections = map[string]indismqgo.Connection{}
	return c
}

// func NewConnection(Connection Connection, conn interface{}, connType uint8) *Connection {
// 	return &Connection{Connection: Connection, Conn: conn}
// }

// func (cc *Connection) Send(m *MsgBuffer) error {
// 	return cc.Connection(m, cc)
// }

func (cs *Connections) Send(key string, m *indismqgo.MsgBuffer) error {
	cc := cs.GetConnection(key)
	if cc == nil {
		return errors.New("No Connection Information Function")
	}
	err := cc.Send(m)
	if err != nil {
		return err
	}
	return nil
}

func (cs *Connections) GetConnection(key string) indismqgo.Connection {
	cs.lock.RLock()
	val := cs.connections[key]
	cs.lock.RUnlock()
	return val
}
func (cs *Connections) SetConnection(key string, val indismqgo.Connection) {
	cs.lock.Lock()
	cs.connections[key] = val
	cs.lock.Unlock()

}
func (cs *Connections) DelConnection(key string) {
	if val := cs.GetConnection(key); val != nil {
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
