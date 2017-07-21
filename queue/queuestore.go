package queue

import (
	"github.com/t4i/indismqgo"
	"sync"
	"time"
)

var _ indismqgo.QueueStore = (*Queues)(nil)

func NewQueues() *Queues {

	qs := &Queues{}
	qs.queues = map[string]*Queue{}
	return qs
}

//Queues collection of multiple queue(s)
type Queues struct {
	sync.RWMutex
	queues map[string]*Queue
}

func (q *Queues) GetQueue(key string) *Queue {
	if queue, ok := q.queues[key]; ok {
		return queue
	}
	return nil
}

func (q *Queues) GetQueueOrNew(key string) *Queue {
	queue := q.GetQueue(key)
	if queue == nil {
		queue = &Queue{}
		q.queues[key] = queue
	}
	return queue
}

func (q *Queues) SetQueue(key string, queue *Queue) {
	q.queues[key] = queue
}

func (q *Queues) DelQueue(key string) {
	delete(q.queues, key)
}
func (q *Queues) ProcessQueue(key string) {

}
func (q *Queues) AckQueue(key string, conn indismqgo.Sender) {

}

// func (q *Queues) PutQueue(msg *indismqgo.MsgBuffer, conn indismqgo.Sender) (queue *Queue) {
// 	q.Lock()
// 	ok := false
// 	queueName := string(msg.Fields.To())
// 	if queue, ok = q.queues[queueName]; ok {
// 		queue.queue = append(queue.queue, &MsgCon{msg: msg, conn: conn})
// 	} else {
// 		queue = &Queue{dequeue: make(map[string]*MsgCon), reply: make(map[string]*MsgCon), container: q}
// 		q.queues[queueName] = queue
// 		queue.queue = append(queue.queue, &MsgCon{msg: msg, conn: conn})
// 	}
// 	q.Unlock()
// 	return queue
// }

// func (q *Queues) PutReply(m *MsgCon) (queue *Queue) {
// 	q.Lock()
// 	ok := false
// 	queueName := string(m.msg.Fields.From())
// 	if queue, ok = q.queues[queueName]; !ok {
// 		queue = &Queue{dequeue: make(map[string]*MsgCon), reply: make(map[string]*MsgCon), container: q}
// 		q.queues[queueName] = queue
// 		// queue.queue = append(queue.queue, msg)
// 	}
// 	queue.reply[string(m.msg.Fields.MsgId())] = m
// 	q.Unlock()
// 	return queue
// }

func (q *Queues) ProcessTimeout() {
	now := time.Now()
	go func() {
		for _, queue := range q.queues {
			for _, v := range queue.queue {
				if v.timeout != nil && v.timeout.After(now) {
					//delete
				}
			}
			for _, v := range queue.reply {
				if v.timeout != nil && v.timeout.After(now) {
					//delete
				}
			}
			for _, v := range queue.dequeue {
				if v.timeout != nil && v.timeout.After(now) {
					//deletes
				}
			}
		}
	}()

}

func (q *Queues) Stats() map[string]QueueStat {
	stats := map[string]QueueStat{}
	for k, queue := range q.queues {
		stat := QueueStat{}
		stats[k] = stat
		for range queue.queue {
			stat.Queue++
		}
		for range queue.reply {
			stat.Reply++
		}
		for range queue.dequeue {
			stat.Dequeue++
		}
	}
	return stats
}
