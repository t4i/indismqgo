package indismqgo

import ()

//Handler
type Handler func(*MsgBuffer, Connection) error

type HandlerStore interface {
	GetHandler(key string) Handler
	SetHandler(key string, val Handler)
	DelHandler(key string)
}
type MessageStore interface {
	GetMessage(key string) (m *MsgBuffer)
	SetMessage(key string, m *MsgBuffer)
	DelMessage(key string)
}

type QueueStore interface {
	ProcessTimeout()
	// PutQueue(msg *MsgBuffer, conn Connection)
	// PopQueue()
	// GetQueue(key string) Queue
	// GetQueueOrNew(key string) Queue
	// DelQueue(key string)
	// AckQueue(key string, conn Connection)
	//SetQueue(key string, q Queue)
	// PutQueue(msg *MsgBuffer, conn Connection) (queue Queue)
	// PutQueueReply(msg *MsgBuffer, conn Connection) (queue Queue)
	// GetQueueReply(m *MsgBuffer) (msg *MsgBuffer, conn Connection, queue Queue)
	ProcessQueue(key string)
}

// type Queue interface {
// 	QueuePut(msg *MsgBuffer, conn Connection)
// 	// QueuePop() interface{}
// 	// QueuePeek() interface{}
// 	//QueueDel(key string)
// 	QueueAck(key string, conn Connection)
// }
type Connection interface {
	Send(m *MsgBuffer) error
}

type ConnectionStore interface {
	GetConnection(key string) Connection
	SetConnection(key string, val Connection)
	DelConnection(key string)
}

type SubscriberStore interface {
	AddSubscriber(client string, path string)
	AddChannel(path string)
	DelChannel(path string)
	DelSubscriber(client string, path string)
	DelSubscriberAll(client string)
	GetSubscriberList(key string) map[string]bool
}
type Subscribers interface {
	ConnectionStore
	SubscriberStore
}

//for context
type OnConnection interface {
	OnConnection(m *MsgBuffer, conn Connection) (ok bool)
}
type OnMessage interface {
	OnMessage(m *MsgBuffer, conn Connection) (ok bool)
}

type OnUnknown interface {
	OnUnknown(m *MsgBuffer, conn Connection) (queue bool)
}

type Subscriber interface {
	OnSubscribe(m *MsgBuffer, conn Connection) (ok bool)
	OnUnsubscribe(m *MsgBuffer, conn Connection)
}
