package broker

import (
	"crypto/tls"
	"github.com/t4i/indismqgo"
	"github.com/t4i/indismqgo/connections"
	"github.com/t4i/indismqgo/handlers"
	"github.com/t4i/indismqgo/messages"
	"github.com/t4i/indismqgo/queue"
	"github.com/t4i/indismqgo/subscribers"
	"net/http"
)

//Server ...
type Broker struct {
	name            []byte
	debug           bool
	GetCertificate  func(hello *tls.ClientHelloInfo) (*tls.Certificate, error)
	HTTPAuthHandler func(h *http.Request) (ok bool, claims map[string]string)
	WsAuthHandler   func(h *http.Request) (ok bool, claims map[string]string)
	*queue.Queues
	*messages.Messages
	*handlers.Handlers
	*connections.Connections
	*subscribers.Subscribers
	//indismqgo.ContextImpl
}

// func (s *Server) NewConnectionType(name string, defaultSender SenderFunc, authHandler AuthHandlerFunc) {
// 	s.ConnectionTypes[name] = &ConnectionType{Name: name, DefaultSender: defaultSender, AuthHandler: authHandler}
// }
var _ indismqgo.QueueStore = (*Broker)(nil)
var _ indismqgo.MessageStore = (*Broker)(nil)
var _ indismqgo.HandlerStore = (*Broker)(nil)
var _ indismqgo.SenderStore = (*Broker)(nil)
var _ indismqgo.Subscribers = (*Broker)(nil)

//NewServer ...
func NewBroker(name string) *Broker {
	broker := &Broker{}
	broker.Name([]byte(name))
	broker.Queues = queue.NewQueues()
	broker.Messages = messages.NewMessages()
	broker.Handlers = handlers.NewHandlers()
	broker.Connections = connections.NewConnections()
	broker.Subscribers = subscribers.NewSubscribers()
	return broker
}
func (c *Broker) Name(set []byte) []byte {
	if set != nil {
		c.name = set
	}
	return c.name
}
func (c *Broker) Debug(set *bool) bool {
	if set != nil {
		c.debug = *set
	}
	return c.debug
}

func (c *Broker) Recieve(m *indismqgo.MsgBuffer, conn indismqgo.Sender) error {
	return indismqgo.Process(c, m, conn)
}

func (c *Broker) RecieveRaw(data []byte, conn indismqgo.Sender) error {
	return indismqgo.ProcessRaw(c, data, conn)
}
func (b *Broker) ProcessNext(q *queue.Queue, conn indismqgo.Sender) {
	// stat := q.Stats()
	// if b.Debug(nil) {
	// 	log.Println(string(b.Name(nil)), q, "queue len", stat.Queue, "reply len", stat.Reply, "dequeue len", stat.Dequeue, "prefetch", stat.Prefetch)
	// }

	// for (q.prefetch > 0 && len(q.dequeue) < q.prefetch) || len(q.queue) > 0 {

	// 	q.processLock.Lock()
	// 	defer q.processLock.Unlock()
	// 	next := q.Peek()
	// 	if q.Handler != nil {

	// 		m := q.Pop()
	// 		q.Handler(m.msg, m.conn)
	// 		return
	// 	}
	// 	//check if is a response
	// 	if reply, replyQueue := queues.GetReply(next.msg); reply != nil {
	// 		m := q.Pop()
	// 		if reply.conn != nil {
	// 			if err := reply.conn.Send(m.msg); err != nil {
	// 				q.RePut(string(m.msg.Fields.MsgId()))
	// 			} else if m.msg.Fields.Guarantee() == indismqgo.GuaranteeAT_LEAST_ONCE {
	// 				//ctx.MakeReply(reply,StatusA)
	// 				delete(replyQueue.reply, string(m.msg.Fields.MsgId()))
	// 				q.RePut(string(m.msg.Fields.MsgId()))
	// 			}

	// 		}
	// 	}

	// 	//check for connection
	// 	// if conn2 := ctx.Connections.Get(string(next.msg.Fields.To())); conn2 != nil {
	// 	// 	m := q.Pop()
	// 	// 	conn2.Send(m.msg)
	// 	// 	ctx.ProcessNext(q, conn)
	// 	// 	return
	// 	// }
	// 	// //check for subscribers
	// 	// subs := ctx.Subscribers.Get(string(next.msg.Fields.To()))
	// 	// for k, v := range subs {
	// 	// 	if v {
	// 	// 		if conn2 := ctx.Connections.Get(k); conn != nil {
	// 	// 			conn2.Send(q.Pop().msg)
	// 	// 			return
	// 	// 		}
	// 	// 	}
	// 	// }

	// }
}

func (b *Broker) OnConnection(m *indismqgo.MsgBuffer, conn indismqgo.Sender) (ok bool) {
	//b.Connections.SetSender(string(m.Fields.From()), conn)
	return true
}

func (b *Broker) OnDisconnected(m *indismqgo.MsgBuffer, conn indismqgo.Sender) (ok bool) {
	b.Connections.DelSender(string(m.Fields.From()))
	b.Subscribers.DelSubscriberAll(string(m.Fields.From()))
	return true
}

func (b *Broker) OnSubscribe(m *indismqgo.MsgBuffer, conn indismqgo.Sender) (ok bool) {
	b.Subscribers.AddSubscriber(string(m.Fields.From()), string(m.Fields.Path()))
	return true
}
func (b *Broker) OnUnSubscribe(m *indismqgo.MsgBuffer, conn indismqgo.Sender) {
	b.Subscribers.DelSubscriber(string(m.Fields.From()), string(m.Fields.Path()))
}
func (b *Broker) OnUnknown(m *indismqgo.MsgBuffer, conn indismqgo.Sender) (ok bool) {

	return true
}
