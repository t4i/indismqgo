package indismqgo

import ()

//Connection interface
//Send Method to be called to send a message using the connection
//ConnectionRequest
type Connection interface {
	Send(m *MsgBuffer) error
}

// func (ctx *Context) NewConnection(isDuplex bool) *Connection {
// 	return &Connection{Context: ctx, Duplex: isDuplex}
// }

// func (ct *Connection) Send(m *MsgBuffer) (bool, error) {
// 	return ct.Sender(m, ct.Ref(conn))
// }

// func (ct *Connection) Recieve(m *MsgBuffer, conn interface{}) error {
// 	return ct.Reciever(m, ct.Ref(conn))
// }

// func (ct *Connection) RecieveRaw(data []byte, conn interface{}) error {
// 	m := ParseMsg(data)
// 	return ct.Reciever(m, ct.Ref(conn))
// }

// func (ct *ConnType) Ref(conn interface{}) *ConnRef {
// 	return &ConnRef{Conn: conn, Type: ct}
// }

// func (ctx *Imq) newConnectionRequest(m *MsgBuffer, conn *Connection, minConnType uint8, onConnect onConnectHandler) (reply *MsgBuffer, connStatus uint8, err error) {

// 	if onConnect != nil {
// 		reply, connStatus, err = onConnect(m, conn, minConnType)
// 	} else {
// 		connStatus = ConnTypeHalfDuplex
// 	}
// 	if reply == nil && connStatus != ConnTypeNone && minConnType < 3 {
// 		reply, err = ctx.Reply(m, StatusOK, nil)
// 		if err != nil {
// 			return nil, ConnTypeNone, err
// 		}
// 	} else if reply == nil && connStatus == ConnTypeNone {
// 		log.Println("connection refused", string(m.Fields.From()))
// 		reply, err = ctx.Reply(m, StatusBadRequest, []byte(StatusText(StatusBadRequest)))
// 		if err != nil {
// 			return nil, ConnTypeNone, err
// 		}
// 	}
// 	return reply, connStatus, err
// }

// func (ctx *Imq) DefaultConnectFullDuplex(m *MsgBuffer, conn *Connection, minConnType uint8) (reply *MsgBuffer, connStatus uint8, err error) {
// 	if m != nil && conn != nil {
// 		conn.ConnType = ConnTypeDuplex
// 		ctx.ConnectionStore.Set(string(m.Fields.From()), conn)
// 		return nil, ConnTypeDuplex, nil
// 	}
// 	return nil, ConnTypeNone, err
// }

// func (ctx *Imq) DefaultUnknownHandler(m *MsgBuffer, conn *Connection) (reply *MsgBuffer, err error) {
// 	if m.Fields.Action() == schema.ActionRESPONSE {
// 		//check if have message
// 		om := ctx.MessageStore.Get(string(m.Fields.MsgId()))
// 		if om != nil && om.Callback != nil {
// 			om.Callback(m, conn)
// 		} else {
// 			reply, err = ctx.Reply(m, StatusNotFound, []byte(StatusText(StatusNotFound)))
// 			if err != nil {
// 				return nil, err
// 			}
// 		}
// 		ctx.MessageStore.Del(string(m.Fields.MsgId()))

// 	} else {
// 		//check if known connection
// 		if ctx.Debug {
// 			log.Println("unknown Req")
// 		}
// 		subConn := ctx.ConnectionStore.Get(string(m.Fields.To()))
// 		if subConn != nil {
// 			subConn.Send(m)
// 			if m.Fields.Callback() != 0 {
// 				if ctx.Debug {
// 					log.Println("callback set")
// 				}
// 				ctx.MessageStore.Set(string(m.Fields.MsgId()), m)
// 				m.Callback = func(r *MsgBuffer, _conn *Connection) (*MsgBuffer, error) {
// 					if ctx.Debug {
// 						log.Println("uknown handler called")
// 					}
// 					conn.Send(r)
// 					return nil, nil
// 				}
// 			}
// 			return nil, nil
// 		}
// 		reply, err = ctx.Reply(m, StatusInternalServerError, []byte(StatusText(StatusInternalServerError)))
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return reply, err

// }
