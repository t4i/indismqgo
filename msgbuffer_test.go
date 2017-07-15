package indismqgo

import (
	"errors"
	"log"
	"testing"
)

var m *MsgBuffer

func init() {
	var err error
	m, err = NewMsgBuffer(
		[]byte("messageid"), 1, 2, []byte("to"), []byte("from"), []byte("path"),
		[]byte("authorization"), []byte("body"), map[string]string{"key": "value"},
		func(msg *MsgBuffer, conn Connection) error {
			if msg != m {
				return errors.New("msg not equal in callback")
			}
			if conn == nil || conn.Send(m) != nil {
				return errors.New("conn not correct in callback")
			}
			return errors.New("success")

		})
	if m == nil {
		log.Fatal("Empty Message during setup")
	}
	if err != nil {
		log.Fatal("Error during setup", err)
	}
}

type testConn struct {
	test string
}

func (tc *testConn) Send(m *MsgBuffer) error {

	return nil
}

func (conn *testConn) Events() *ConnEvents{
	return nil
}
func (conn *testConn) OnConnect(key string) {

}
func (conn *testConn) OnDisconnect(key string) {

}
func TestNewMsgBuffer(t *testing.T) {
	//check fields
	if string(m.Fields.MsgId()) != "messageid" {
		t.Error("Wrong Message Id")
	}
	if m.Fields.Action() != 1 {
		t.Error("Wrong Action")
	}
	if m.Fields.Status() != 2 {
		t.Error("Wrong Status")
	}
	if string(m.Fields.To()) != "to" {
		t.Error("Wrong To")
	}
	if string(m.Fields.From()) != "from" {
		t.Error("Wrong From")
	}
	if string(m.Fields.Path()) != "path" {
		t.Error("Wrong PAth")
	}
	if string(m.Fields.Authorization()) != "authorization" {
		t.Error("Wrong Authorization")
	}
	if string(m.Fields.BodyBytes()) != "body" {
		t.Error("Wrong Body")
	}
	if m.Fields.MetaLength() != 1 {
		t.Error("Meta not set")
	} else {
		meta := new(KeyVal)
		if !m.Fields.Meta(meta, 0) {
			t.Error("Error retrieving meta")
		} else {
			if string(meta.Key()) != "key" || string(meta.Value()) != "value" {
				t.Error("Wrong Meta")
			}
		}
	}

	if m.Callback(m, &testConn{test: "testConn"}).Error() != "success" {
		t.Error("Wrong callback")
	}
	meta := m.Meta()
	if meta == nil {
		t.Error("No Meta")
	} else if val, ok := meta["key"]; !ok {
		t.Error("Invalid Key")
	} else if val != "value" {
		t.Error("Invalid Value")
	}

}

func TestParseMessage(t *testing.T) {
	m2 := ParseMsg(m.Data, nil)

	if string(m.Data) != string(m2.Data) {
		t.Error("Parse not equal")
	}
}

func TestToObject(t *testing.T) {
	o := m.ToObject()
	if o.MsgID != "messageid" {
		t.Error("Wrong Message Id")
	}
	if o.Action != 1 {
		t.Error("Wrong Action")
	}
	if o.Status != 2 {
		t.Error("Wrong Status")
	}
	if o.To != "to" {
		t.Error("Wrong To")
	}
	if o.From != "from" {
		t.Error("Wrong From")
	}
	if o.Path != "path" {
		t.Error("Wrong PAth")
	}
	if o.Authorization != "authorization" {
		t.Error("Wrong Authorization")
	}
	if string(o.Body) != "body" {
		t.Error("Wrong Body")
	}

	if o.Callback(m, &testConn{test: "testConn"}).Error() != "success" {
		t.Error("Wrong callback")
	}

	if o.Meta == nil {
		t.Error("No Meta")
	} else if val, ok := o.Meta["key"]; !ok {
		t.Error("Invalid Key")
	} else if val != "value" {
		t.Error("Invalid Value")
	}
}
