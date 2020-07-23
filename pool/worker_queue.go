package pool

import "errors"

type queue interface {
	enqueue(worker *Worker) error
	dequeue() *Worker
	len() int
	isEmpty() bool
	clear()
}

type workerQueue struct {
	items []*Worker
	size  int
}

func newWorkerQueue(size int) *workerQueue {
	return &workerQueue{
		items: make([]*Worker, 0, size),
		size:  size,
	}
}

func (q *workerQueue) enqueue(worker *Worker) error {
	if worker == nil {
		return errors.New("nil pointer of worker")
	}
	q.items = append(q.items, worker)
	return nil
}

func (q *workerQueue) dequeue() *Worker {
	l := q.len()

	if l == 0 {
		return nil
	}

	w := q.items[0]
	q.items = q.items[1:l]

	return w
}

func (q *workerQueue) len() int {
	return len(q.items)
}

func (q *workerQueue) isEmpty() bool {
	return q.len() == 0
}

func (q *workerQueue) clear() {
	for i := 0; i < q.len(); i++ {
		q.items[i].sendTask(nil)
	}
	q.items =q.items[:0]
}
