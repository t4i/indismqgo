package indismqgo

import (
	"fmt"
	fb "github.com/google/flatbuffers/go"
	"github.com/t4i/indismqgo/schema/IndisMQ"
)

//Msg ... Imq.Msg.rawData Imq.Msg.
type MsgBuffer struct {
	Data      []byte
	Fields    *IndisMQ.Imq
	Callback  Handler
	Timestamp int64
	meta      map[string]string
	Context   Context
}

//Meta read only access to metadata
func (m *MsgBuffer) Meta() map[string]string {
	if m.meta != nil || len(m.meta) > 0 {
		return m.meta
	} else if m.Fields.MetaLength() > 0 {
		parseMeta(m)
		return m.meta
	} else {
		return nil
	}
}

func parseMeta(m *MsgBuffer) {
	if l := m.Fields.MetaLength(); l > 0 {
		m.meta = make(map[string]string)
		for i := 0; i < l; i++ {
			meta := new(IndisMQ.KeyVal)
			if m.Fields.Meta(meta, i) {
				m.meta[string(meta.Key())] = string(meta.Value())
			}
		}
	}
}

func (m *MsgBuffer) ToObject() *MsgObject {
	o := &MsgObject{
		MsgID:         string(m.Fields.MsgId()),
		Action:        m.Fields.Action(),
		Status:        m.Fields.Status(),
		To:            string(m.Fields.To()),
		From:          string(m.Fields.From()),
		Path:          string(m.Fields.Path()),
		Authorization: string(m.Fields.Authorization()),
		Body:          m.Fields.BodyBytes(),
	}
	parseMeta(m)
	o.Meta = m.meta
	o.Callback = m.Callback
	return o
}

func NewMsgBuffer(msgId []byte, action int8, status uint16, to []byte, from []byte, path []byte, authorization []byte, body []byte, meta map[string]string, callback Handler) (*MsgBuffer, error) {
	m := &MsgBuffer{}
	var msgIDOffset, toOffset, fromOffset, pathOffset, authorizationOffset, bodyOffset, metaOffset fb.UOffsetT
	//var metaOffset []fb.UOffsetT
	builder := fb.NewBuilder(0)
	if msgId != nil && len(msgId) > 0 {
		msgIDOffset = builder.CreateByteString(msgId)
	} else {
		msgIDOffset = builder.CreateByteString(NewRandID())
	}
	if to != nil && len(to) > 0 {
		toOffset = builder.CreateByteString(to)
	}
	if from != nil && len(from) > 0 {
		fromOffset = builder.CreateByteString(from)
	}
	if path != nil && len(path) > 0 {
		pathOffset = builder.CreateByteString(path)
	}
	if authorization != nil && len(authorization) > 0 {
		authorizationOffset = builder.CreateByteString(authorization)
	}
	if body != nil && len(body) > 0 {
		bodyOffset = builder.CreateByteVector(body)
	}

	if l := len(meta); meta != nil && l > 0 {
		var metaOffsets []fb.UOffsetT
		for k, v := range meta {
			keyOffset := builder.CreateString(k)
			valOffset := builder.CreateString(v)
			IndisMQ.KeyValStart(builder)
			IndisMQ.KeyValAddKey(builder, keyOffset)
			IndisMQ.KeyValAddValue(builder, valOffset)
			metaOffsets = append(metaOffsets, IndisMQ.KeyValEnd(builder))
		}
		l := len(metaOffsets)
		IndisMQ.ImqStartMetaVector(builder, l)
		for i := 0; i < l; i++ {
			builder.PrependUOffsetT(metaOffsets[i])
		}
		metaOffset = builder.EndVector(l)
	}
	IndisMQ.ImqStart(builder)
	IndisMQ.ImqAddMsgId(builder, msgIDOffset)
	IndisMQ.ImqAddAction(builder, action)
	IndisMQ.ImqAddStatus(builder, status)
	IndisMQ.ImqAddTo(builder, toOffset)
	IndisMQ.ImqAddFrom(builder, fromOffset)
	IndisMQ.ImqAddPath(builder, pathOffset)
	IndisMQ.ImqAddAuthorization(builder, authorizationOffset)
	IndisMQ.ImqAddBody(builder, bodyOffset)
	IndisMQ.ImqAddMeta(builder, metaOffset)
	if callback != nil {
		IndisMQ.ImqAddCallback(builder, 1)
		m.Callback = callback
	}
	rpc := IndisMQ.ImqEnd(builder)
	builder.Finish(rpc)
	buf := builder.FinishedBytes()
	m.Data = buf
	m.Fields = IndisMQ.GetRootAsImq(buf, 0)
	return m, nil
}

//ParseMsg ...
func ParseMsg(data []byte, ctx Context) (m *MsgBuffer) {
	if data == nil {
		return nil
	}
	m = &MsgBuffer{}
	m.Fields = IndisMQ.GetRootAsImq(data, 0)
	m.Data = data
	m.Context = ctx
	return
}

func (m *MsgBuffer) String() string {
	return fmt.Sprintln(m.ToObject(), string(m.Fields.BodyBytes()))
	//fmt.Println("Recieved ID:", string(m.Fields.MsgId()), "From:", string(m.Fields.From()), " To:", string(m.Fields.To()),  EnumNamesMsgType[int(m.Fields.MsgType())], " ",  EnumNamesSts[int(m.Fields.Sts())], " ",  EnumNamesCmd[int(m.Fields.Cmd())], string(m.Fields.StsMsg()))
}
