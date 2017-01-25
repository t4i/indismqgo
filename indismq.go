package Imq

import (
	"fmt"

	"github.com/dchest/uniuri"
	fb "github.com/google/flatbuffers/go"
	schema "github.com/t4i/indismqgo/schema/IndisMQ"
)

//NewImq ...
func NewImq() *Imq {
	var i = new(Imq)
	i.Debug = false
	i.Handlers = make(map[string]Handler)
	i.Messages = make(map[string]*Msg)
	i.Subscribers = make(map[string]map[string]bool)
	i.OnReady = func(client string) {}
	i.name = []byte("unnamed")
	return i
}

//Imq ...
type Imq struct {
	Handlers      map[string]Handler
	Debug         bool
	name          []byte
	BrokerHandler Handler
	RelayHandler  Handler
	Messages      map[string]*Msg
	Subscribers   map[string]map[string]bool
	OnReady       func(client string)
	Handler
	Msg
}

// //NewAuth ...
// func NewAuth(user *string, pass *string, token *string) (auth *Auth) {
// 	auth = &Auth{}
// 	builder := fb.NewBuilder(1024)
// 	userOffset := builder.CreateString(*user)
// 	passOffset := builder.CreateString(*pass)
// 	tokenOffset := builder.CreateString(*token)
// 	schema.AuthStart(builder)
// 	schema.AuthAddUser(builder, userOffset)
// 	schema.AuthAddPass(builder, passOffset)
// 	schema.AuthAddToken(builder, tokenOffset)
// 	a := schema.AuthEnd(builder)
// 	builder.Finish(a)
// 	buf := builder.FinishedBytes()
// 	auth.Data = buf
// 	auth.Fields = schema.GetRootAsAuth(buf, 0)
// 	return
// }

//Handler ...
type Handler func(m *Msg) *Msg

//RPCSender ...
//type RPCSender func(data *[]byte) bool

//Msg ... Imq.Msg.rawData Imq.Msg.
type Msg struct {
	Data     *[]byte
	Fields   *schema.Imq
	Callback Handler
}

// var statusLock = &sync.RWMutex{}
// var handlerLock = &sync.RWMutex{}
// var senderLock = &sync.RWMutex{}
// var subLock = &sync.RWMutex{}

//SetBrokerHandler ...
func (i *Imq) SetBrokerHandler(handler Handler) {
	i.BrokerHandler = handler
}

//SetRelayHandler ...
func (i *Imq) SetRelayHandler(handler Handler) {
	i.RelayHandler = handler
}

//ParseMsg ...
func ParseMsg(data *[]byte) (m *Msg) {
	if data == nil {
		return nil
	}
	m = &Msg{}
	m.Fields = schema.GetRootAsImq(*data, 0)
	m.Data = data
	return
}

func (i *Imq) getImqMessage(id string) (imqMessage *Msg) {
	var present bool
	//statusLock.RLock()
	if imqMessage, present = i.Messages[id]; !present {
		return nil
	}
	//statusLock.RUnlock()
	return
}

//GetMessagesize ...
func (i *Imq) GetMessagesize() int {
	return len(i.Messages)
}

//DelMessage ...
func (i *Imq) DelMessage(id string) {
	//statusLock.Lock()
	var present bool
	if _, present = i.Messages[id]; present {
		delete(i.Messages, id)
	}
	//statusLock.Unlock()
}

//AddSubscriber ...
func (i *Imq) AddSubscriber(client string, path string) {
	var present bool

	//subLock.Lock()
	if _, present = i.Subscribers[path]; !present {
		i.Subscribers[path] = make(map[string]bool)
	}
	i.Subscribers[path][client] = false
	//subLock.Unlock()
}

//DelSubscriber ...
func (i *Imq) DelSubscriber(client string, path string) {
	var present, present2 bool
	//subLock.Lock()
	if _, present = i.Subscribers[path]; present {
		if _, present2 = i.Subscribers[path][client]; present2 {
			delete(i.Subscribers[path], client)
		}
		if len(i.Subscribers[path]) < 1 {
			delete(i.Subscribers, path)
		}
	}

	//subLock.Unlock()
}

//DelSubscriberAll ...
func (i *Imq) DelSubscriberAll(client string) {

	for k := range i.Subscribers {
		i.DelSubscriber(client, k)
	}
}

//SetHandler ...
func (i *Imq) SetHandler(path string, handler Handler) {
	//handlerLock.Lock()
	i.Handlers[path] = handler
	//handlerLock.Unlock()
}

//GetHandler ...
func (i *Imq) GetHandler(path string) (handler Handler) {
	//handlerLock.RLock()
	handler = i.Handlers[path]
	//handlerLock.RUnlock()
	return
}

//DelHandler ...
func (i *Imq) DelHandler(path string) {
	//handlerLock.Lock()
	delete(i.Handlers, path)
	//handlerLock.Unlock()
}

//Syn ...
func (i *Imq) Syn(stsMsg string, callback Handler, user string) *Msg {
	uuid, uuidErr := newUID()
	if uuidErr != nil {
		fmt.Printf("error: %v\n", uuidErr)
	}

	m := i.makeImq(uuid, i.name, nil, 0, nil, schema.MsgTypeCMD, schema.StsREQ, schema.CmdSYN, []byte(stsMsg), -1, nil, callback, []byte(user))
	if callback != nil {
		i.Messages[string(uuid)] = m
	}
	return m
}

//Err ...
func (i *Imq) Err(m *Msg, stsMsg string, err int8, user string) *Msg {
	return i.makeImq(m.Fields.MsgId(), i.name, m.Fields.From(), m.Fields.Broker(), m.Fields.Path(), m.Fields.MsgType(), schema.StsERROR, -1, []byte(stsMsg), err, nil, nil, []byte(user))
}

//Ready ...
func (i *Imq) Ready(to string, dest string, callback Handler, user string) *Msg {
	//encode
	uuid, uuidErr := newUID()
	if uuidErr != nil {
		fmt.Printf("error: %v\n", uuidErr)
	}
	m := i.makeImq(uuid, i.name, []byte(to), 0, []byte(dest), schema.MsgTypeCMD, schema.StsREQ, schema.CmdREADY, nil, -1, nil, callback, []byte(user))
	if callback != nil {
		i.Messages[string(uuid)] = m
	}
	return m
}

//Success ...
func (i *Imq) Success(m *Msg, stsMsg string, user string) *Msg {
	return i.makeImq(m.Fields.MsgId(), i.name, m.Fields.From(), m.Fields.Broker(), m.Fields.Path(), m.Fields.MsgType(), schema.StsSUCCESS, m.Fields.Cmd(), []byte(stsMsg), -1, nil, nil, []byte(user))
}

//Req ...
func (i *Imq) Req(to string, dest string, msg []byte, callback Handler, user string) *Msg {
	//encode
	uuid, uuidErr := newUID()
	if uuidErr != nil {
		fmt.Printf("error: %v\n", uuidErr)
	}
	m := i.makeImq(uuid, i.name, []byte(to), 0, []byte(dest), schema.MsgTypeSINGLE, schema.StsREQ, -1, nil, -1, msg, callback, []byte(user))
	if callback != nil {
		i.Messages[string(uuid)] = m
	}
	return m
}

//Rep ...
func (i *Imq) Rep(m *Msg, stsMsg string, msg []byte, user string) *Msg {
	return i.makeImq(m.Fields.MsgId(), i.name, m.Fields.From(), m.Fields.Broker(), m.Fields.Path(), m.Fields.MsgType(), schema.StsREP, m.Fields.Cmd(), []byte(stsMsg), -1, msg, nil, []byte(user))
}

//Sub ...
func (i *Imq) Sub(to string, path string, handler Handler, callback Handler, user string) *Msg {
	//make and return a subscribe message
	//fmt.Println("Path", path)
	uuid, uuidErr := newUID()
	if uuidErr != nil {
		fmt.Printf("error: %v\n", uuidErr)
	}
	if handler != nil {
		i.SetHandler(path, handler)
	}
	m := i.makeImq(uuid, i.name, []byte(to), 0, []byte(path), schema.MsgTypeCMD, schema.StsREQ, schema.CmdSUB, nil, -1, nil, callback, []byte(user))
	if callback != nil {
		i.Messages[string(uuid)] = m
	}
	return m
}

//UnSub ...
func (i *Imq) UnSub(to string, path string, callback Handler, user string) *Msg {
	//make and return a unsubscribe message
	uuid, uuidErr := newUID()
	if uuidErr != nil {
		fmt.Printf("error: %v\n", uuidErr)
	}
	i.DelHandler(path)
	m := i.makeImq(uuid, i.name, []byte(to), 0, []byte(path), schema.MsgTypeCMD, schema.StsREQ, schema.CmdUNSUB, nil, -1, nil, callback, []byte(user))
	if callback != nil {
		i.Messages[string(uuid)] = m
	}
	return m
}

//BrokerReplay ...
func (i *Imq) BrokerReplay(m *Msg, handler func(string, *Msg), callback Handler) {
	var r *Msg
	if m.Fields.MsgType() == schema.MsgTypeCAST {
		r = i.makeImq(m.Fields.MsgId(), m.Fields.From(), nil, 0, m.Fields.Path(), schema.MsgTypeCAST, schema.StsREQ, -1, m.Fields.StsMsg(), -1, m.Fields.BodyBytes(), callback, m.Fields.User())
		i.sendMult(r, handler)
	} else {
		r = i.makeImq(m.Fields.MsgId(), m.Fields.From(), nil, 0, m.Fields.Path(), schema.MsgTypeQUEUE, schema.StsREQ, -1, m.Fields.StsMsg(), -1, m.Fields.BodyBytes(), callback, m.Fields.User())
		i.sendQueue(r, handler)
	}

}

//Mult ...
func (i *Imq) Mult(broker bool, path string, msg []byte, handler func(string, *Msg), callback Handler, user string) *Msg {
	//make a pub request and call a closure
	uuid, uuidErr := newUID()
	if uuidErr != nil {
		fmt.Printf("error: %v\n", uuidErr)
	}
	hasBroker := byte(0)
	if broker {
		hasBroker = 1
	}
	m := i.makeImq(uuid, i.name, nil, hasBroker, []byte(path), schema.MsgTypeCAST, schema.StsREQ, -1, nil, -1, msg, callback, []byte(user))
	if !broker {
		i.sendMult(m, handler)
		return nil
	}
	if callback != nil {
		i.Messages[string(uuid)] = m
	}
	return m

}
func (i *Imq) sendMult(m *Msg, handler func(string, *Msg)) {
	for k := range i.Subscribers[string(m.Fields.Path())] {
		handler(k, m)
	}
}

//Queue ...
func (i *Imq) Queue(broker bool, path string, msg []byte, handler func(string, *Msg), callback Handler, user string) *Msg {
	uuid, uuidErr := newUID()
	if uuidErr != nil {
		fmt.Printf("error: %v\n", uuidErr)
	}
	hasBroker := byte(0)
	if broker {
		hasBroker = 1
	}
	m := i.makeImq(uuid, i.name, nil, hasBroker, []byte(path), schema.MsgTypeQUEUE, schema.StsREQ, -1, nil, -1, msg, callback, []byte(user))
	if !broker {
		i.sendQueue(m, handler)
		return nil
	}
	if callback != nil {
		i.Messages[string(uuid)] = m
	}
	return m

}
func (i *Imq) sendQueue(m *Msg, handler func(string, *Msg)) {
	success := false
	takeNext := false
	path := string(m.Fields.Path())
	if _, ok := i.Subscribers[path]; ok {
		if len(i.Subscribers[path]) > 0 {

			for k, v := range i.Subscribers[path] {

				if takeNext {
					handler(k, m)
					i.Subscribers[path][k] = true
					success = true
					//fmt.Println("takeNext ", k)
				}

				if v {
					takeNext = true
					i.Subscribers[path][k] = false
				}
			}
			if !success {
				for k := range i.Subscribers[path] {
					handler(k, m)
					i.Subscribers[path][k] = true
					//fmt.Println("!success ", k)
					return
				}

			}

		}
	}

}
func (i *Imq) makeImq(id []byte, from []byte, to []byte, broker byte, path []byte, msgType int8, sts int8, cmd int8, stsMsg []byte, err int8, body []byte, callback Handler, user []byte) (m *Msg) {
	if i.Debug {
		fmt.Println("Send id:", string(id), "to", string(to), " from", string(from), " type ", schema.EnumNamesMsgType[int(msgType)], " sts ", schema.EnumNamesSts[int(sts)], " cmd ", schema.EnumNamesCmd[int(cmd)])
	}
	m = &Msg{}
	builder := fb.NewBuilder(0)
	imqID := builder.CreateByteString(id)
	var bodyOffset = builder.CreateByteVector(body)
	var stsMsgOffset = builder.CreateByteString(stsMsg)
	pathOffset := builder.CreateByteString(path)
	fromOffset := builder.CreateByteString(from)
	toOffset := builder.CreateByteString(to)
	userOffset := builder.CreateByteString(user)
	schema.ImqStart(builder)
	schema.ImqAddUser(builder, userOffset)
	schema.ImqAddFrom(builder, fromOffset)
	schema.ImqAddTo(builder, toOffset)
	schema.ImqAddMsgId(builder, imqID)
	if broker == 1 {
		schema.ImqAddBroker(builder, 1)
	}
	if callback != nil {
		schema.ImqAddCallback(builder, 1)
		m.Callback = callback
	}
	schema.ImqAddPath(builder, pathOffset)
	schema.ImqAddSts(builder, sts)
	if body != nil {
		schema.ImqAddBody(builder, bodyOffset)
	}
	if stsMsg != nil {
		schema.ImqAddStsMsg(builder, stsMsgOffset)
	}
	if err != -1 {
		schema.ImqAddErr(builder, err)
	}
	if msgType != -1 {
		schema.ImqAddMsgType(builder, msgType)
	}
	if cmd != -1 {
		schema.ImqAddCmd(builder, cmd)
	}
	rpc := schema.ImqEnd(builder)
	builder.Finish(rpc)
	buf := builder.FinishedBytes()

	//decide if we should add to msg queue
	m.Data = &buf
	m.Fields = schema.GetRootAsImq(buf, 0)
	return
}

//RecieveRawMessage ...
func (i *Imq) RecieveRawMessage(data *[]byte, user string) (reply *Msg) {
	//decode RPC
	if data == nil {
		return
	}
	m := ParseMsg(data)
	reply = i.RecieveMessage(m, user)
	if i.Debug {
		fmt.Println("RecieveRaw From ", string(reply.Fields.From()))
	}
	return
}

//RecieveMessage ...
func (i *Imq) RecieveMessage(m *Msg, user string) (reply *Msg) {
	//fmt.Println("recieved", m)

	if m == nil {
		//return error?
		return nil
	}
	if i.Debug {
		fmt.Println("Recieved ID:", string(m.Fields.MsgId()), "From:", string(m.Fields.From()), " To:", string(m.Fields.To()), schema.EnumNamesMsgType[int(m.Fields.MsgType())], " ", schema.EnumNamesSts[int(m.Fields.Sts())], " ", schema.EnumNamesCmd[int(m.Fields.Cmd())], string(m.Fields.StsMsg()))
	}
	if m.Fields.Broker() == 1 {
		//call broker handler
		if i.BrokerHandler != nil {
			reply = i.BrokerHandler(m)
		} else {
			fmt.Println("not a broker")
			reply = i.Err(m, "not a broker", schema.ErrNO_HANDLER, user)
		}
	} else if len(m.Fields.To()) > 0 && string(m.Fields.To()) != string(i.name) {
		//call relay handler
		if i.RelayHandler != nil {
			reply = i.RelayHandler(m)
		} else {
			fmt.Println("not a relay")
			reply = i.Err(m, "not a relay", schema.ErrNO_HANDLER, user)
		}
	} else if m.Fields.MsgType() == schema.MsgTypeCMD {
		reply = i.handleCmd(m, user)
	} else if m.Fields.Sts() == schema.StsREQ {
		if handler := i.GetHandler(string(m.Fields.Path())); handler != nil {
			reply = handler(m)
		}
	} else { //its a reply
		Imq := i.getImqMessage(string(m.Fields.MsgId()))
		if Imq != nil && Imq.Callback != nil {
			Imq.Callback(m)
		}
		i.DelMessage(string(m.Fields.MsgId()))
	}

	return
}
func (i *Imq) handleCmd(Imq *Msg, user string) (m *Msg) {
	if Imq.Fields.Sts() == schema.StsREQ {
		switch Imq.Fields.Cmd() {
		case schema.CmdSUB:
			i.AddSubscriber(string(Imq.Fields.From()), string(Imq.Fields.Path()))
			m = i.Success(Imq, "", user)
		case schema.CmdSYN:
			m = i.Success(Imq, "", user)
		case schema.CmdUNSUB:
			i.DelSubscriber(string(Imq.Fields.From()), string(Imq.Fields.Path()))
			m = i.Success(Imq, "", user)
		case schema.CmdREADY:
			fmt.Println("ready")
			m = i.Success(Imq, "", user)
			i.OnReady(string(Imq.Fields.From()))
		default:
		}
	} else {
		msg := i.getImqMessage(string(Imq.Fields.MsgId()))
		if msg != nil && msg.Callback != nil {
			msg.Callback(Imq)
		}
		i.DelMessage(string(Imq.Fields.MsgId()))
	}
	return
}

//SetName ...
func (i *Imq) SetName(newName string) {
	//should notify all connections
	i.name = []byte(newName)
}

// newUUID generates a random UUID according to RFC 4122
func newUID() ([]byte, error) {
	// uuid := make([]byte, 16)
	// n, err := io.ReadFull(rand.Reader, uuid)
	// if n != len(uuid) || err != nil {
	// 	return nil, err
	// }
	// // variant bits; see section 4.1.1
	// uuid[8] = uuid[8]&^0xc0 | 0x80
	// // version 4 (pseudo-random); see section 4.1.3
	// uuid[6] = uuid[6]&^0xf0 | 0x40
	// return []byte(fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])), nil
	return []byte(uniuri.New()), nil
}

/*
ideas
-allow combination of broker and relay so you can relay to a broker
-allow a dns like lookup of to by relay

*/
