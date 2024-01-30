package queue

import "log"

type FuncQueue struct {
	QueueFunc func(...any)
	Args      []any
}

var funqueue chan *FuncQueue
var funqueue_inited bool

func Init(queueLen int64) {
	funqueue_inited = true
	funqueue = make(chan *FuncQueue, queueLen)
	go funcAutoRun()
}

func funcAutoRun() {
	defer func() {
		err := recover()
		log.Print(err)
		go funcAutoRun()
	}()
	for {
		funcQueue := <-funqueue
		funcQueue.QueueFunc(funcQueue.Args...)
	}
}

func NewFuncQueue(f func(...any), args ...any) {
	if !funqueue_inited {
		Init(102400)
	}
	funcQueue := &FuncQueue{
		QueueFunc: f,
		Args:      args,
	}
	funqueue <- funcQueue
}
