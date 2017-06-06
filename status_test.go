package indismqgo

import (
	"testing"
)

func TestStatusText(t *testing.T) {

	if s := StatusText(StatusOK); s != "OK" {
		t.Error("Incorrect StatusOK text", s)
	}
}
