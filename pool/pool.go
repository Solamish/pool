package pool

import (
	"sync"
	"sync/atomic"
)

type sig struct{}

type Pool struct {
	// capacity of the pool.
	capacity int32

	// running is the number of the currently running goroutines.
	running int32

	// state represents the state of pool.
	state int32

	// freeSignal is used to notice pool there are available
	// workers which can be sent to work.
	freeSignal chan sig

	// workers is a queue that stores the available workers.
	workers queue

	workerCache sync.Pool

	lock sync.Mutex

}

func NewPool(size int) (*Pool, error){
	if size <= 0 {
		return nil, ErrPoolInitialized
	}

	p := &Pool{
		capacity:    int32(size),
		freeSignal:  make(chan sig),
		workers:     newWorkerQueue(size),
	}
	p.workerCache.New = func() interface{} {
		return &Worker{
			pool: p,
			task: make(chan func(), workerChanCap),
		}
	}

	return p, nil
}

// Running returns the number of the currently running goroutines.
func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// Cap returns the capacity of this pool.
func (p *Pool) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

// Left returns the available goroutines to work.
func (p *Pool) Left() int {
	return p.Cap() - p.Running()
}

// Submit submits a task to this pool.
func (p *Pool) Submit(task func()) error{
	if atomic.LoadInt32(&p.state) == CLOSED {
		return ErrPoolClosed
	}
	w := p.getWorker()
	w.sendTask(task)
	return nil
}

// stop this worker.
func (p *Pool) Stop() {
		atomic.StoreInt32(&p.state,CLOSED)
		p.lock.Lock()
		p.workers.clear()
		p.lock.Unlock()
}

// --------------------------------------------------------

// incRunning increases the number of the currently running goroutines.
func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

// decRunning decreases the number of the currently running goroutines.
func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

func (p *Pool) getWorker() *Worker {
	var w *Worker
	spawnWorker := func() {
		w = p.workerCache.Get().(*Worker)
		w.run()
	}
	waiting := false
	p.lock.Lock()

	w = p.workers.dequeue()
	// 如果没有空闲的worker
	if w == nil {
		// 如果pool已经满了，阻塞等待到pool有空闲worker
		if p.Running() >= p.Cap() {
			waiting = true
		// 如果pool没有满，就新建一个worker
		} else {
			spawnWorker()
		}
	}

	p.lock.Unlock()

	if waiting {
		// 阻塞，直到收到worker空闲的通知
		<-p.freeSignal
		for {
			p.lock.Lock()
			w = p.workers.dequeue()

			if p.Running() == 0 {
				p.lock.Unlock()
				spawnWorker()
				return w
			}

			if w == nil {
				p.lock.Unlock()
				continue
			}
			p.lock.Unlock()
			break
		}

	}
	return w

}

func (p *Pool)putWorker(worker *Worker) bool{
	p.lock.Lock()
	err := p.workers.enqueue(worker)

	if err != nil {
		p.lock.Unlock()
		return false
	}
	p.lock.Unlock()
	// TODO Notice
	p.freeSignal <- sig{}
	return true
}
