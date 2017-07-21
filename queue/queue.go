package queue

import (
	"github.com/t4i/indismqgo"
	"sync"
	"time"
)

type Queue struct {
	sync.RWMutex
	Handler     indismqgo.Handler
	queue       []*MsgCon
	dequeue     map[string]*MsgCon
	reply       map[string]*MsgCon
	prefetch    int
	container   *Queues
	processLock sync.Mutex
}

//MsgCon a struct with a MsgBuffer and Connection related to it
type MsgCon struct {
	msg     *indismqgo.MsgBuffer
	conn    indismqgo.Sender
	timeout *time.Time
}

func (q *Queue) QueuePut(msg *indismqgo.MsgBuffer, conn indismqgo.Sender) {
	q.Lock()
	q.queue = append(q.queue, &MsgCon{msg: msg, conn: conn})
	q.Unlock()
	return
}

func (q *Queue) Stats() *QueueStat {
	stats := &QueueStat{}
	stats.Queue = len(q.queue)
	stats.Reply = len(q.reply)
	stats.Dequeue = len(q.dequeue)
	stats.Prefetch = q.prefetch
	return stats
}

func (q *Queue) QueuePop() *MsgCon {
	q.Lock()
	defer q.Unlock()
	var m *MsgCon
	m, q.queue = q.queue[len(q.queue)-1], q.queue[:len(q.queue)-1]
	if m != nil && m.msg != nil && m.msg.Fields.Guarantee() == indismqgo.GuaranteeAT_LEAST_ONCE {
		q.dequeue[string(m.msg.Fields.MsgId())] = m
	}
	if m.msg != nil && m.msg.Fields.Callback() == 1 {
		// /	q.container.PutReply(m)
	}

	return m

}

func (q *Queue) QueuePeek() *MsgCon {
	q.RLock()
	defer q.RUnlock()
	if len(q.queue) > 0 {
		return q.queue[0]
	}
	return nil

}
func (q *Queue) QueueReput(msgId string) {
	queued := q.dequeue[msgId]
	delete(q.dequeue, msgId)
	q.queue = append(q.queue, queued)

}
func (q *Queue) QueueAck(msgId string, replyConn indismqgo.Sender) {
	queued := q.dequeue[msgId]
	if queued.msg.Fields.Callback() == 1 {
		q.reply[msgId] = queued
	}
	delete(q.dequeue, msgId)
}

func (q *Queues) GetReply(m *indismqgo.MsgBuffer) (*MsgCon, *Queue) {

	if queue, ok := q.queues[string(m.Fields.To())]; ok {
		if reply, ok := queue.reply[string(m.Fields.MsgId())]; ok {
			return reply, queue
		}
	}
	return nil, nil
}
