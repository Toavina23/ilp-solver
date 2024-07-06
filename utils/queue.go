package utils

type Queue[T any] struct {
	items []T
}

// NewQueue creates and returns a new queue
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{items: make([]T, 0)}
}

// Enqueue adds an item to the back of the queue
func (q *Queue[T]) Enqueue(item T) {
	q.items = append(q.items, item)
}

// Dequeue removes and returns the front item from the queue
func (q *Queue[T]) Dequeue() (T, bool) {
	var zero T
	if len(q.items) == 0 {
		return zero, false
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item, true
}

// Front returns the front item without removing it
func (q *Queue[T]) Front() (T, bool) {
	var zero T
	if len(q.items) == 0 {
		return zero, false
	}
	return q.items[0], true
}

// IsEmpty checks if the queue is empty
func (q *Queue[T]) IsEmpty() bool {
	return len(q.items) == 0
}

// Size returns the number of items in the queue
func (q *Queue[T]) Size() int {
	return len(q.items)
}
