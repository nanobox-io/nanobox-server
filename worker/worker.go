package worker

import (
	"sync"

	"github.com/nanobox-core/nanobox-server/config"
)

// structs
type (

	//
	Worker struct {
		sync.Mutex
		sync.WaitGroup

		queue []Job
	}
)

// interfaces
type (

	//
	Job interface {
		Process()
	}
)

//
func New() *Worker {
	return &Worker{queue: []Job{}}
}

//
func (w *Worker) Queue(job Job) {
	w.Lock()
	w.queue = append(w.queue, job)
	w.Unlock()
}

//
func (w *Worker) QueueAndProcess(job Job) {
	w.Queue(job)
	w.Process()
}

//
func (w *Worker) QueueAndProcessNow(job Job) {
	w.Queue(job)
	w.ProcessNow()
}

//
func (w *Worker) Process() {
	go w.ProcessNow()
}

//
func (w *Worker) ProcessNow() {
	config.Log.Debug("[NANOBOX :: WORKER] Proccessing job...\n")

	//
	w.Wait()
	w.Add(1)

	//
	for {
		if job, ok := w.nextJob(); ok {
			w.processJob(job)
		} else {
			break
		}
	}

	w.Done()
}

// private

//
func (w *Worker) nextJob() (Job, bool) {
	var job Job

	w.Lock()
	if len(w.queue) >= 1 {
		job, w.queue = w.queue[0], w.queue[1:]
		return job, true
	}
	w.Unlock()

	return nil, false
}

//
func (w *Worker) processJob(job Job) {

	//
	defer func() {
		if err := recover(); err != nil {
			config.Log.Error("[NANOBOX :: WORKER] Job failed: %+v\n", err)
		}
	}()

	//
	job.Process()
}
