package utils

import (
	"context"
	"sync"
	"time"
)

type Lock interface {
	sync.Locker
	TryLock() bool
}

type lockWithTimeout struct {
	m    sync.Mutex
	t    time.Duration
	c    func()
	done chan any
}

func (l *lockWithTimeout) Lock() {
	l.m.Lock()
	l.acquireLock()
}

func (l *lockWithTimeout) TryLock() bool {
	if !l.m.TryLock() {
		return false
	}

	l.acquireLock()

	return true
}

func (l *lockWithTimeout) Unlock() {
	l.c()
	<-l.done
}

func (l *lockWithTimeout) acquireLock() {
	ctx, cancel := context.WithTimeout(context.Background(), l.t)
	l.c = cancel
	l.done = make(chan any)

	go func() {
		<-ctx.Done()
		l.c = func() {}
		l.m.Unlock()
		close(l.done)
	}()
}

func NewLockWithTimeout(t time.Duration) Lock {
	done := make(chan any)
	close(done)

	return &lockWithTimeout{
		t:    t,
		c:    func() {},
		done: done,
	}
}
