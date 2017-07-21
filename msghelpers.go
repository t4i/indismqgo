package indismqgo

import ()

func MakeCtxReply(ctx Context, m *MsgBuffer, statusCode uint16, body []byte) (*MsgBuffer, error) {
	reply, err := NewMsgBuffer(m.Fields.MsgId(), ActionRESPONSE, statusCode, m.Fields.From(), ctx.Name(nil), nil, nil, body, nil, nil)
	if err != nil {
		return nil, err
	}
	//reply.Conn = m.Conn
	reply.Context = ctx
	return reply, nil
}

func NewCtxConnectionMsg(ctx Context, to []byte, authorization []byte, callback Handler) (*MsgBuffer, error) {
	return NewMsgBuffer(nil, ActionCONNECT, 0, to, ctx.Name(nil), nil, authorization, nil, nil, callback)
}

func NewCtxMsgObject(ctx Context, to string, action int8, path string, body []byte, callback Handler) *MsgObject {
	return &MsgObject{MsgID: string(NewRandID()), Action: action, To: to, Path: path, From: string(ctx.Name(nil)), Body: body, Callback: callback}
}
