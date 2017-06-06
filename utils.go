package indismqgo

import (
	"github.com/dchest/uniuri"
	"github.com/satori/go.uuid"
	"sync"
	"time"
)

// newRandID generates a random UUID according to RFC 4122
func NewRandID() []byte {
	return []byte(uniuri.New())
}

func NewUUID() []byte {
	return uuid.NewV4().Bytes()
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
