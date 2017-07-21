package broker

import (
	"github.com/t4i/indismqgo"
)

func (b *Broker) NewConnectionMsg(to []byte, authorization []byte, callback indismqgo.Handler) (*indismqgo.MsgBuffer, error) {
	return indismqgo.NewCtxConnectionMsg(b, to, authorization, callback)
}

func (b *Broker) MakeReply(m *indismqgo.MsgBuffer, statusCode uint16, body []byte) (*indismqgo.MsgBuffer, error) {
	return indismqgo.MakeCtxReply(b, m, statusCode, body)
}

func (b *Broker) NewMsgObject(to string, action int8, path string, body []byte, callback indismqgo.Handler) *indismqgo.MsgObject {
	return indismqgo.NewCtxMsgObject(b, to, action, path, body, callback)
}
