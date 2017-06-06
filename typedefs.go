package indismqgo

import ()

//Handler ...
type Handler func(*MsgBuffer, Connection) error

// type ConnClass int8

// type Sender func(m *MsgBuffer, conn *Connection) (bool, error)
// type Reciever func(m *MsgBuffer, conn *Connection) error
