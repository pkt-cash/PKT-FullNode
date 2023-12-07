package mailbox

import "sync"

type Mailbox[T any] struct {
	m        sync.Mutex
	val      T
	waitChan chan struct{}
}

func NewMailbox[T any](initialVal T) Mailbox[T] {
	return Mailbox[T]{
		val:      initialVal,
		waitChan: make(chan struct{}),
	}
}

func (mb *Mailbox[T]) Load() T {
	mb.m.Lock()
	t := mb.val
	mb.m.Unlock()
	return t
}

func (mb *Mailbox[T]) Store(t T) {
	mb.m.Lock()
	mb.val = t
	close(mb.waitChan)
	mb.waitChan = make(chan struct{})
	mb.m.Unlock()
}

func (mb *Mailbox[T]) AwaitUpdate() T {
	mb.m.Lock()
	ch := mb.waitChan
	mb.m.Unlock()
	<-ch
	mb.m.Lock()
	t := mb.val
	mb.m.Unlock()
	return t
}
