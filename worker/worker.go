package worker

import (
	"sync"

	"github.com/nanobox-core/nanobox-server/config"
)

// structs
type (

	//
	Worker struct {
		sync.WaitGroup

		Concurrent bool
		doTex sync.Mutex
		queueTex sync.Mutex
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
	return &Worker{
		Concurrent: false,
		doTex: sync.Mutex{},
		queueTex: sync.Mutex{},
		queue: []Job{},
	}
}

//
func (w *Worker) Queue(job Job) {
	w.queueTex.Lock()
	defer w.queueTex.Unlock()

	w.queue = append(w.queue, job)
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

	w.doTex.Lock()
	defer w.doTex.Unlock()
	//
	w.Add(1)

	//
	for {
		if job, ok := w.nextJob(); ok {
			w.Add(1)
			w.processJob(job)
		} else {
			break
		}
	}

	w.Done()
	w.Wait()
}

// private

//
func (w *Worker) nextJob() (Job, bool) {
	w.queueTex.Lock()
	defer w.queueTex.Unlock()

	var job Job

	if len(w.queue) >= 1 {
		job, w.queue = w.queue[0], w.queue[1:]
		return job, true
	}

	return nil, false
}

//
func (w *Worker) processJob(job Job) {
	defer w.Done()
	//
	defer func() {
		if err := recover(); err != nil {
			config.Log.Error("[NANOBOX :: WORKER] Job failed: %+v\n", err)
		}
	}()

	//
	job.Process()
}
