package indismqgo

import ()

type MsgObject struct {
	MsgID         string
	Action        int8
	Status        uint16
	To            string
	From          string
	Path          string
	Authorization string
	Body          []byte
	Meta          map[string]string
	Callback      Handler
}

func NewMsgObject(msgId string, action int8, status uint16, to string, from string, path string, auth string, body []byte, meta map[string]string, callback Handler) *MsgObject {
	return &MsgObject{MsgID: msgId, Action: action, Status: status, To: to, Path: path, From: from, Authorization: auth, Meta: meta, Body: body, Callback: callback}
}

func (o *MsgObject) ToBuffer() (*MsgBuffer, error) {
	return NewMsgBuffer([]byte(o.MsgID), o.Action, o.Status, []byte(o.To), []byte(o.From), []byte(o.Path), []byte(o.Authorization), o.Body, o.Meta, o.Callback)
}


