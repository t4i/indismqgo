// automatically generated by the FlatBuffers compiler, do not modify

package indismqgo

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type Imq struct {
	_tab flatbuffers.Table
}

func GetRootAsImq(buf []byte, offset flatbuffers.UOffsetT) *Imq {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Imq{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Imq) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Imq) MsgId() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Imq) Action() int8 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetInt8(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Imq) MutateAction(n int8) bool {
	return rcv._tab.MutateInt8Slot(6, n)
}

func (rcv *Imq) Status() uint16 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetUint16(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Imq) MutateStatus(n uint16) bool {
	return rcv._tab.MutateUint16Slot(8, n)
}

func (rcv *Imq) To() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Imq) From() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Imq) Path() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Imq) Authorization() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(16))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Imq) Callback() byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(18))
	if o != 0 {
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Imq) MutateCallback(n byte) bool {
	return rcv._tab.MutateByteSlot(18, n)
}

func (rcv *Imq) Body(j int) byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(20))
	if o != 0 {
		a := rcv._tab.Vector(o)
		return rcv._tab.GetByte(a + flatbuffers.UOffsetT(j*1))
	}
	return 0
}

func (rcv *Imq) BodyLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(20))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *Imq) BodyBytes() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(20))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Imq) Meta(obj *KeyVal, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(22))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *Imq) MetaLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(22))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *Imq) Guarantee() int8 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(24))
	if o != 0 {
		return rcv._tab.GetInt8(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Imq) MutateGuarantee(n int8) bool {
	return rcv._tab.MutateInt8Slot(24, n)
}

func (rcv *Imq) Timeout() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(26))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Imq) MutateTimeout(n int32) bool {
	return rcv._tab.MutateInt32Slot(26, n)
}

func ImqStart(builder *flatbuffers.Builder) {
	builder.StartObject(12)
}
func ImqAddMsgId(builder *flatbuffers.Builder, MsgId flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(MsgId), 0)
}
func ImqAddAction(builder *flatbuffers.Builder, Action int8) {
	builder.PrependInt8Slot(1, Action, 0)
}
func ImqAddStatus(builder *flatbuffers.Builder, Status uint16) {
	builder.PrependUint16Slot(2, Status, 0)
}
func ImqAddTo(builder *flatbuffers.Builder, To flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(3, flatbuffers.UOffsetT(To), 0)
}
func ImqAddFrom(builder *flatbuffers.Builder, From flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(4, flatbuffers.UOffsetT(From), 0)
}
func ImqAddPath(builder *flatbuffers.Builder, Path flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(5, flatbuffers.UOffsetT(Path), 0)
}
func ImqAddAuthorization(builder *flatbuffers.Builder, Authorization flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(6, flatbuffers.UOffsetT(Authorization), 0)
}
func ImqAddCallback(builder *flatbuffers.Builder, Callback byte) {
	builder.PrependByteSlot(7, Callback, 0)
}
func ImqAddBody(builder *flatbuffers.Builder, Body flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(8, flatbuffers.UOffsetT(Body), 0)
}
func ImqStartBodyVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(1, numElems, 1)
}
func ImqAddMeta(builder *flatbuffers.Builder, Meta flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(9, flatbuffers.UOffsetT(Meta), 0)
}
func ImqStartMetaVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func ImqAddGuarantee(builder *flatbuffers.Builder, Guarantee int8) {
	builder.PrependInt8Slot(10, Guarantee, 0)
}
func ImqAddTimeout(builder *flatbuffers.Builder, Timeout int32) {
	builder.PrependInt32Slot(11, Timeout, 0)
}
func ImqEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}