package indismqgo

import ()

//Handler
type Handler func(*MsgBuffer, Sender) error

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
	// PutQueue(msg *MsgBuffer, conn Sender)
	// PopQueue()
	// GetQueue(key string) Queue
	// GetQueueOrNew(key string) Queue
	// DelQueue(key string)
	// AckQueue(key string, conn Sender)
	//SetQueue(key string, q Queue)
	// PutQueue(msg *MsgBuffer, conn Sender) (queue Queue)
	// PutQueueReply(msg *MsgBuffer, conn Sender) (queue Queue)
	// GetQueueReply(m *MsgBuffer) (msg *MsgBuffer, conn Sender, queue Queue)
	ProcessQueue(key string)
}

// type Queue interface {
// 	QueuePut(msg *MsgBuffer, conn Sender)
// 	// QueuePop() interface{}
// 	// QueuePeek() interface{}
// 	//QueueDel(key string)
// 	QueueAck(key string, conn Sender)
// }
type Sender interface {
	Send(m *MsgBuffer) error
}

type SenderStore interface {
	GetSender(key string) Sender
	SetSender(key string, val Sender)
	DelSender(key string)
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
	SenderStore
	SubscriberStore
}

//for context
type OnConnection interface {
	OnConnection(m *MsgBuffer, conn Sender) (ok bool)
}
type OnMessage interface {
	OnMessage(m *MsgBuffer, conn Sender) (ok bool)
}

type OnUnknown interface {
	OnUnknown(m *MsgBuffer, conn Sender) (queue bool)
}

type Subscriber interface {
	OnSubscribe(m *MsgBuffer, conn Sender) (ok bool)
	OnUnsubscribe(m *MsgBuffer, conn Sender)
}
