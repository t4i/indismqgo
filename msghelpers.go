package indismqgo

import ()

func (ctx *Context) MakeReply(m *MsgBuffer, statusCode uint16, body []byte) (*MsgBuffer, error) {
	reply, err := NewMsgBuffer(m.Fields.MsgId(), ActionRESPONSE, statusCode, m.Fields.From(), ctx.Name, nil, nil, body, nil, nil)
	if err != nil {
		return nil, err
	}
	//reply.Conn = m.Conn
	reply.Context = ctx
	return reply, nil

	return nil, nil
}

func (ctx *Context) NewConnectionMsg(to []byte, authorization []byte, callback Handler) (*MsgBuffer, error) {
	return NewMsgBuffer(nil, ActionCONNECT, 0, to, ctx.Name, nil, authorization, nil, nil, callback)
}
