package pool

import (
	"errors"
	"runtime"
)

const (
	// ACTIVE represents that the pool is opened.
	ACTIVE = iota

	// CLOSED represents that the pool is closed.
	CLOSED
)

var (

	// ErrPoolClosed will be returned when submitting task to a closed pool.
	ErrPoolClosed = errors.New("this pool has been closed")

	// ErrPoolInitialized will be returned when initializing pool illegally.
	ErrPoolInitialized = errors.New("the size of this pool must greater than zero")

	workerChanCap = func() int {
		// Use blocking workerChan if GOMAXPROCS=1.
		// This immediately switches Serve to WorkerFunc, which results
		// in higher performance (under go1.5 at least).
		if runtime.GOMAXPROCS(0) == 1 {
			return 0
		}

		// Use non-blocking workerChan if GOMAXPROCS>1,
		// since otherwise the Serve caller (Acceptor) may lag accepting
		// new connections if WorkerFunc is CPU-bound.
		return 1
	}()

)