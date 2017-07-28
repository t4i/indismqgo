package indismqgo

import (
	"errors"
	"log"
	"reflect"
	"testing"
)

var mo *MsgObject

func init() {

	mo = NewMsgObject(
		"messageid", 1, 2, "to", "from", "path",
		"authorization", []byte("body"), map[string]string{"key": "value"},
		func(msg *MsgBuffer, conn Connection) error {
			if msg != m {

				return errors.New("msg not equal in callback")
			}
			if conn == nil || conn.Send(m) != nil {
				return errors.New("conn not correct in callback")
			}
			return errors.New("success")

		})
	if mo == nil {
		log.Fatal("Empty Message during setup")
	}

}

func TestNewMsgObject(t *testing.T) {
	if mo.MsgID != "messageid" {
		t.Error("Wrong Message Id")
	}
	if mo.Action != 1 {
		t.Error("Wrong Action")
	}
	if mo.Status != 2 {
		t.Error("Wrong Status")
	}
	if mo.To != "to" {
		t.Error("Wrong To")
	}
	if mo.From != "from" {
		t.Error("Wrong From")
	}
	if mo.Path != "path" {
		t.Error("Wrong PAth")
	}
	if mo.Authorization != "authorization" {
		t.Error("Wrong Authorization")
	}
	if string(mo.Body) != "body" {
		t.Error("Wrong Body")
	}

	if m.Callback(m, &testConn{test: "testConn"}).Error() != "success" {
		t.Error("Wrong callback")
	}
	if mo.Meta == nil {
		t.Error("No Meta")
	} else if val, ok := mo.Meta["key"]; !ok {
		t.Error("Invalid Key")
	} else if val != "value" {
		t.Error("Invalid Value")
	}
}

func TestToBuffer(t *testing.T) {
	buf, err := mo.ToBuffer()
	if err != nil {
		t.Error(err)
	}
	if string(buf.Data) != string(m.Data) {
		t.Error("Buffer not equal")
	}
}
func TestMsgObjectCompatibility(t *testing.T) {

	mo2 := m.ToObject

	if reflect.DeepEqual(mo, mo2) {
		t.Error("MsgObject not equal to MsgBuffer")
	}

}
