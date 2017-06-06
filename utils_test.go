package indismqgo

import (
	"strings"
	"testing"
)

func TestNewRandID(t *testing.T) {
	id := NewRandID()
	if len(id) != 16 {
		t.Error("Not 16 Characters")
	}
}

func TestNewRandIdUnique(t *testing.T) {
	id := NewRandID()
	id2 := NewRandID()
	if strings.Compare(string(id), string(id2)) == 0 {
		t.Error("Not Unique")
	}
}

func TestNewUUID(t *testing.T) {
	id := NewUUID()
	if len(id) != 16 {
		t.Error("Not 16 Characters")
	}
}

func TestNewUUIDUnique(t *testing.T) {
	id := NewUUID()
	id2 := NewUUID()
	if strings.Compare(string(id), string(id2)) == 0 {
		t.Error("Not Unique")
	}
}

var result []byte

func BenchmarkUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result = NewUUID()
	}
}
func BenchmarkRandID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result = NewRandID()
	}
}
