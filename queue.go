package indismqgo

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Queue struct {
	sync.RWMutex
	Handler     Handler
	queue       []*MsgCon
	dequeue     map[string]*MsgCon
	reply       map[string]*MsgCon
	prefetch    int
	container   *Queues
	processLock sync.Mutex
}

//MsgCon a struct with a MsgBuffer and Connection related to it
type MsgCon struct {
	msg     *MsgBuffer
	conn    Connection
	timeout *time.Time
}

//Queues collection of multiple queue(s)
type Queues struct {
	sync.RWMutex
	queues map[string]*Queue
}

type QueueItems struct {
	queue  *Queue
	delete []string
	rePut  []string
}

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
					//delete
				}
			}
		}
	}()

}

func (q *Queues) GetQueue(key string) *Queue {
	if queue, ok := q.queues[key]; ok {
		return queue
	}
	return nil
}

func (q *Queues) Put(msg *MsgBuffer, conn Connection) (queue *Queue) {
	q.Lock()
	ok := false
	queueName := string(msg.Fields.To())
	if queue, ok = q.queues[queueName]; ok {
		queue.queue = append(queue.queue, &MsgCon{msg: msg, conn: conn})
	} else {
		queue = &Queue{dequeue: make(map[string]*MsgCon), reply: make(map[string]*MsgCon), container: q}
		q.queues[queueName] = queue
		queue.queue = append(queue.queue, &MsgCon{msg: msg, conn: conn})
	}
	q.Unlock()
	return queue
}

func (q *Queues) PutReply(m *MsgCon) (queue *Queue) {
	q.Lock()
	ok := false
	queueName := string(m.msg.Fields.From())
	if queue, ok = q.queues[queueName]; !ok {
		queue = &Queue{dequeue: make(map[string]*MsgCon), reply: make(map[string]*MsgCon), container: q}
		q.queues[queueName] = queue
		// queue.queue = append(queue.queue, msg)
	}
	queue.reply[string(m.msg.Fields.MsgId())] = m
	q.Unlock()
	return queue
}

func (q *Queue) Put(msg *MsgBuffer, conn Connection) {
	q.Lock()
	q.queue = append(q.queue, &MsgCon{msg: msg, conn: conn})
	q.Unlock()
	return
}

// func (q *Queues) Pop(key string, timeout int) (msg *MsgBuffer) {
// 	q.Lock()
// 	if val, ok := q.queues[string(msg.Fields.To())]; ok && len(val.messages) > 0 {
// 		return val.Pop(timeout)
// 	}
// 	q.Unlock()
// 	return nil
// }
type QueueStat struct {
	queue   int
	dequeue int
	reply   int
}

func (q *Queues) Stats() string {
	stats := map[string]QueueStat{}
	for k, queue := range q.queues {
		stat := QueueStat{}
		stats[k] = stat
		for range queue.queue {
			stat.queue++
		}
		for range queue.reply {
			stat.reply++
		}
		for range queue.dequeue {
			stat.dequeue++
		}
	}
	return fmt.Sprintln(stats)
}

func (q *Queue) Pop() *MsgCon {
	q.Lock()
	defer q.Unlock()
	var m *MsgCon
	m, q.queue = q.queue[len(q.queue)-1], q.queue[:len(q.queue)-1]
	if m != nil && m.msg != nil && m.msg.Fields.Guarantee() == GuaranteeAT_LEAST_ONCE {
		q.dequeue[string(m.msg.Fields.MsgId())] = m
	}
	if m.msg != nil && m.msg.Fields.Callback() == 1 {
		q.container.PutReply(m)
	}

	return m

}

func (q *Queue) Peek() *MsgCon {
	q.RLock()
	defer q.RUnlock()
	if len(q.queue) > 0 {
		return q.queue[0]
	}
	return nil

}
func (q *Queue) RePut(msgId string) {
	queued := q.dequeue[msgId]
	delete(q.dequeue, msgId)
	q.queue = append(q.queue, queued)

}
func (q *Queue) Ack(msgId string, replyConn Connection) {
	queued := q.dequeue[msgId]
	if queued.msg.Fields.Callback() == 1 {
		q.reply[msgId] = queued
	}
	delete(q.dequeue, msgId)
}

func (q *Queues) GetReply(m *MsgBuffer) (*MsgCon, *Queue) {

	if queue, ok := q.queues[string(m.Fields.To())]; ok {
		if reply, ok := queue.reply[string(m.Fields.MsgId())]; ok {
			return reply, queue
		}
	}
	return nil, nil
}
func (ctx *Context) ProcessNext(q *Queue, conn Connection) {

	if ctx.Debug {
		log.Println(string(ctx.Name), q, "queue len", len(q.queue), "reply len", len(q.reply), "dequeue len", len(q.dequeue), "prefetch", q.prefetch)
	}

	for (q.prefetch > 0 && len(q.dequeue) < q.prefetch) || len(q.queue) > 0 {

		q.processLock.Lock()
		defer q.processLock.Unlock()
		next := q.Peek()
		if q.Handler != nil {
			m := q.Pop()
			q.Handler(m.msg, m.conn)
			return
		}
		//check if is a response
		if reply, replyQueue := ctx.Queues.GetReply(next.msg); reply != nil {
			m := q.Pop()
			if reply.conn != nil {
				if err := reply.conn.Send(m.msg); err != nil {
					q.RePut(string(m.msg.Fields.MsgId()))
				} else if m.msg.Fields.Guarantee() == GuaranteeAT_LEAST_ONCE {
					//ctx.MakeReply(reply,StatusA)
					delete(replyQueue.reply, string(m.msg.Fields.MsgId()))
					q.RePut(string(m.msg.Fields.MsgId()))
				}

			}
		}

		//check for connection
		if conn2 := ctx.Connections.Get(string(next.msg.Fields.To())); conn2 != nil {
			m := q.Pop()
			conn2.Send(m.msg)
			ctx.ProcessNext(q, conn)
			return
		}
		//check for subscribers
		subs := ctx.Subscribers.Get(string(next.msg.Fields.To()))
		for k, v := range subs {
			if v {
				if conn2 := ctx.Connections.Get(k); conn != nil {
					conn2.Send(q.Pop().msg)
					return
				}
			}
		}

	}
}
