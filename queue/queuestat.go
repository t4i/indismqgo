package queue

import (
	"fmt"
)

type QueueStat struct {
	Queue    int
	Dequeue  int
	Reply    int
	Prefetch int
}

func (q *QueueStat) String() string {
	return fmt.Sprintln(q)
}
