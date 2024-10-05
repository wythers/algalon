package utils

import (
	"sync"
)

type Queuer[T any] interface {
	Enqueue(item *T) error
	Dequeue() (*T, error)
	Suspend() ([]*T, error)

	IsEmpty() bool
	Counter() int
	IsSuspended() bool
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		items:     make([]*T, 0, 128),
		suspended: false,
	}
}

type Queue[T any] struct {
	lock  sync.Mutex
	items []*T

	suspended bool
}

func (q *Queue[T]) Enqueue(item *T) error {
	return q.enqueue(item)
}

func (q *Queue[T]) enqueue(item *T) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.suspended || item == nil {
		return ErrSuspended
	}

	q.items = append(q.items, item)
	return nil
}

func (q *Queue[T]) Dequeue() (*T, error) {
	return q.dequeue()
}

func (q *Queue[T]) dequeue() (*T, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.suspended {
		return nil, ErrSuspended
	}

	if len(q.items) == 0 {
		return nil, ErrEOF
	}

	item := q.items[0]
	q.items = q.items[1:]

	return item, nil
}

func (q *Queue[T]) Suspend() ([]*T, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.suspended {
		return nil, ErrSuspended
	}

	q.suspended = true

	tmp := q.items
	q.items = make([]*T, 0, 128)

	return tmp, nil
}

func (q *Queue[T]) BatchIn(items []*T) error {
	return q.batchIn(items)
}

func (q *Queue[T]) BatchOut() ([]*T, error) {
	return q.batchOut()
}

func (q *Queue[T]) batchIn(items []*T) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.suspended {
		return ErrSuspended
	}

	q.items = append(q.items, items...)
	return nil
}

func (q *Queue[T]) batchOut() ([]*T, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.suspended {
		return nil, ErrSuspended
	}

	if len(q.items) == 0 {
		return nil, ErrEOF
	}

	tmp := q.items
	q.items = make([]*T, 0, 128)

	return tmp, nil
}

func (q *Queue[T]) IsEmpty() bool {
	defer q.lock.Unlock()
	q.lock.Lock()

	return len(q.items) == 0
}

func (q *Queue[T]) IsSuspended() bool {
	defer q.lock.Unlock()
	q.lock.Lock()

	return q.suspended
}

func (q *Queue[T]) Counter() int {
	defer q.lock.Unlock()
	q.lock.Lock()

	return len(q.items)
}
