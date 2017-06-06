package indismqgo

import (
	"bytes"
	"errors"
	"log"
)

func (ctx *Context) RecieveRaw(data []byte, conn Connection) error {
	m := ParseMsg(data, ctx)
	return ctx.Recieve(m, conn)
}

func (ctx *Context) Recieve(m *MsgBuffer, conn Connection) error {
	if m == nil {
		//return error?
		return errors.New("Empty Message ")
	}
	// if m.Conn == nil {
	// 	if conn != nil {
	// 		m.Conn = conn
	// 	} else {
	// 		return errors.New("No Connection specified")
	// 	}
	// }
	if m.Context == nil {
		m.Context = ctx
	}
	if ctx.Debug {
		log.Println(string(ctx.Name), "Recieved", m.String())
	}
	if len(m.Fields.To()) > 0 && string(m.Fields.To()) != string(ctx.Name) {
		if ctx.Debug {
			log.Println("unknown to field")
		}
		if m.Fields.Status() == StatusAccepted {
			if queue := ctx.Queues.GetQueue(string(m.Fields.To())); queue != nil {
				queue.Ack(string(m.Fields.MsgId()), conn)
				ctx.ProcessNext(queue, conn)
			}
		} else {
			queue := ctx.Queues.Put(m, conn)
			ctx.ProcessNext(queue, conn)

		}

		return nil
	}
	switch m.Fields.Action() {
	case ActionRESPONSE:
		om := ctx.Messages.Get(string(m.Fields.MsgId()))
		if om != nil && om.Callback != nil {
			return om.Callback(m, conn)
		}
		ctx.Messages.Del(string(m.Fields.MsgId()))
		return nil
	case ActionSUBSCRIBE:
		if ctx.ConnectionHandler != nil {
			if err := ctx.ConnectionHandler(m, conn); err != nil {
				ctx.MakeReply(m, StatusBadRequest, []byte(err.Error()))
				return err
			}
		}
		subConn := ctx.Connections.Get(string(m.Fields.From()))
		if subConn == nil {
			ctx.Connections.Set(string(m.Fields.From()), conn)
		}
		ctx.Subscribers.AddSubscriber(string(m.Fields.From()), string(m.Fields.Path()))

		r, err := ctx.MakeReply(m, StatusOK, nil)
		if err != nil {
			return err
		}
		err = conn.Send(r)
		if queue := ctx.Queues.GetQueue(string(m.Fields.Path())); queue != nil {
			ctx.ProcessNext(queue, conn)
		}
		return err
	case ActionUNSUBSCRIBE:
		ctx.Subscribers.DelSubscriber(string(m.Fields.From()), string(m.Fields.Path()))
		r, err := ctx.MakeReply(m, StatusOK, nil)
		if err != nil {
			return err
		}
		conn.Send(r)
	case ActionCONNECT:
		if ctx.ConnectionHandler != nil {
			if err := ctx.ConnectionHandler(m, conn); err != nil {
				ctx.MakeReply(m, StatusBadRequest, []byte(err.Error()))
				return err
			}
		}
		ctx.Connections.Set(string(m.Fields.From()), conn)

		r, err := ctx.MakeReply(m, StatusOK, nil)
		if err != nil {
			return err
		}
		err = conn.Send(r)
		if queue := ctx.Queues.GetQueue(string(m.Fields.From())); queue != nil {
			ctx.ProcessNext(queue, conn)
		}
		return err
	case ActionCAST:
		var buf bytes.Buffer
		if handler := ctx.Handlers.Get(string(m.Fields.Path())); handler != nil {
			err := handler(m, conn)
			if err != nil {
				buf.WriteString(err.Error())
			}
		}
		subs := ctx.Subscribers.Get(string(m.Fields.Path()))
		for k := range subs {
			subConn := ctx.Connections.Get(k)
			if subConn != nil {
				err := subConn.Send(m)
				if err != nil {
					buf.WriteString(err.Error())
				}
			}
		}
		if buf.Len() > 0 {
			return errors.New(buf.String())
		}
		return nil
	default:
		if handler := ctx.Handlers.Get(string(m.Fields.Path())); handler != nil {
			err := handler(m, conn)
			if err != nil {
				return err
			}
		} else {
			r, err := ctx.MakeReply(m, StatusNotFound, []byte(StatusText(StatusNotFound)))
			if err != nil {
				return err
			}
			conn.Send(r)
		}
	}
	return nil
}
