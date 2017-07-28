package indismqgo

import (
	"bytes"
	"errors"
	"log"
	"reflect"
)

type Context interface {
	Recieve(m *MsgBuffer, conn Connection) error
	RecieveRaw(data []byte, conn Connection) error
	Name([]byte) []byte
	Debug(*bool) bool
	//ProcessQueue(q *Queue, conn Connection)
}

type ContextImpl struct {
	//Context
	name  []byte
	debug bool
}

func (c *ContextImpl) Name(set []byte) []byte {
	if set != nil {
		c.name = set
	}
	return c.name
}
func (c *ContextImpl) Debug(set *bool) bool {
	if set != nil {
		c.debug = *set
	}
	return c.debug
}

func (c *ContextImpl) Recieve(m *MsgBuffer, conn Connection) error {
	return Process(c, m, conn)
}

func (c *ContextImpl) RecieveRaw(data []byte, conn Connection) error {
	return ProcessRaw(c, data, conn)
}

func ProcessRaw(ctx Context, data []byte, conn Connection) error {
	m := ParseMsg(data, ctx)
	return Process(ctx, m, conn)
}
func Process(ctx Context, m *MsgBuffer, conn Connection) error {
	if m == nil {
		//return error?
		return errors.New("Empty Message ")
	}
	if m.Context == nil {
		m.Context = ctx
	}
	if ctx.Debug(nil) {
		log.Println(string(ctx.Name(nil)), "Recieved", m.String())
	}
	if onMes, ok := ctx.(OnMessage); ok {
		if !onMes.OnMessage(m, conn) {
			if m, err := MakeCtxReply(ctx, m, StatusUnauthorized, nil); err == nil {
				return conn.Send(m)
			} else {
				return err
			}
			return nil
		}
	}
	if len(m.Fields.To()) > 0 && string(m.Fields.To()) != string(ctx.Name(nil)) {
		if ctx.Debug(nil) {
			log.Println("unknown to field")
		}
		impl := false
		//check unknown handler
		if unk, ok := ctx.(OnUnknown); ok {
			unk.OnUnknown(m, conn)
			impl = true
		}
		//Check known connections
		if c, ok := ctx.(ConnectionStore); ok {
			if send := c.GetConnection(string(m.Fields.To())); send != nil {
				return send.Send(m)
			}
		}
		// if q, ok := ctx.(QueueStore); ok {
		// 	if m.Fields.Status() == StatusAccepted {
		// 		if queue := q.GetQueue(string(m.Fields.To())); queue != nil {
		// 			queue.QueueAck(string(m.Fields.MsgId()), conn)
		// 			q.ProcessNext(queue, conn)
		// 		}
		// 	} else {
		// 		queue := q.GetQueueOrNew(string(m.Fields.To()))
		// 		q.ProcessNext(queue, conn)

		// 	}
		// 	impl = true
		// }
		if !impl {
			return sendNotImplemented(ctx, m, conn, "Unknown")
		}
		return nil
	}
	switch m.Fields.Action() {
	case ActionRESPONSE:
		if res, ok := ctx.(MessageStore); ok {
			om := res.GetMessage(string(m.Fields.MsgId()))
			if om != nil && om.Callback != nil {
				return om.Callback(m, conn)
			}
			res.DelMessage(string(m.Fields.MsgId()))
			return nil
		} else {
			log.Println(reflect.TypeOf(ctx))
			return sendNotImplemented(ctx, m, conn, "Messagestore dawg")
		}

	case ActionSUBSCRIBE:
		if s, ok := ctx.(Subscriber); ok {
			if !s.OnSubscribe(m, conn) {
				if m, err := MakeCtxReply(ctx, m, StatusUnauthorized, nil); err == nil {
					return conn.Send(m)
				} else {
					return err
				}
			}
			// if q, ok := ctx.(QueueStore); ok {
			// 	if queue := q.GetQueue(string(m.Fields.Path())); queue != nil {
			// 		q.ProcessNext(queue, conn)
			// 	}
			// }

			return nil
		}
		return sendNotImplemented(ctx, m, conn, "Subscribers")
	case ActionUNSUBSCRIBE:
		if s, ok := ctx.(Subscriber); ok {
			s.OnUnsubscribe(m, conn)
			// if q, ok := ctx.(QueueStore); ok {
			// 	if queue := q.GetQueue(string(m.Fields.Path())); queue != nil {
			// 		q.ProcessNext(queue, conn)
			// 	}
			// }

			return nil
		}

		return sendNotImplemented(ctx, m, conn, "Subscribers")
	case ActionCONNECT:
		impl := false
		if c, ok := ctx.(OnConnection); ok {
			if !c.OnConnection(m, conn) {
				if m, err := MakeCtxReply(ctx, m, StatusUnauthorized, nil); err == nil {
					return conn.Send(m)
				} else {
					return err
				}
			}
			impl = true
		}
		if c, ok := ctx.(ConnectionStore); ok {
			c.SetConnection(string(m.Fields.From()), conn)
			impl = true
		}
		// if q, ok := ctx.(QueueStore); ok {
		// 	if queue := q.GetQueue(string(m.Fields.Path())); queue != nil {
		// 		q.ProcessNext(queue, conn)
		// 	}
		// 	impl = true
		// }
		if !impl {
			return sendNotImplemented(ctx, m, conn, "Connections")
		}
		return nil
	case ActionCAST:
		impl := false
		var buf bytes.Buffer
		if h, ok := ctx.(HandlerStore); ok {
			if handler := h.GetHandler(string(m.Fields.Path())); handler != nil {
				err := handler(m, conn)
				if err != nil {
					buf.WriteString(err.Error())
				}
			}
			impl = true
		}
		if s, ok := ctx.(Subscribers); ok {
			subs := s.GetSubscriberList(string(m.Fields.Path()))
			for k := range subs {
				subConn := s.GetConnection(k)
				if subConn != nil {
					err := subConn.Send(m)
					if err != nil {
						buf.WriteString(err.Error())
					}
				}
			}
			impl = true
		}
		if buf.Len() > 0 {
			return errors.New(buf.String())
		}
		if !impl {
			return sendNotImplemented(ctx, m, conn, "Handlerstore and Subscribers")
		}
		return nil
	default:
		if h, ok := ctx.(HandlerStore); ok {
			if handler := h.GetHandler(string(m.Fields.Path())); handler != nil {
				err := handler(m, conn)
				if err != nil {
					return err
				}
			} else {
				r, err := MakeCtxReply(ctx, m, StatusNotFound, []byte("Handler Not Found"))
				if err != nil {
					return err
				}
				return conn.Send(r)
			}
			return nil
		}
		return sendNotImplemented(ctx, m, conn, "HandlerStore")
	}

}

func sendNotImplemented(ctx Context, m *MsgBuffer, conn Connection, msg string) error {
	r, _ := MakeCtxReply(ctx, m, StatusNotImplemented, []byte(msg+" not implemented"))
	return conn.Send(r)
}
