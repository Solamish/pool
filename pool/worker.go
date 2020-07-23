package pool

type Worker struct {
	pool *Pool
	task chan func()
}

func (w *Worker) sendTask(task func()) {
	w.task <- task
}

func (w *Worker) run() {
	w.pool.incRunning()
	go func() {
		defer func() {
			w.pool.decRunning()
			w.pool.workerCache.Put(w)
		}()

		for f := range w.task {
			if f == nil {
				return
			}
			f()
			if ok := w.pool.putWorker(w); !ok {
				return
			}
		}
	}()
}


