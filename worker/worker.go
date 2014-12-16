package worker

import (
	"fmt"
	"sync"

	"github.com/nanobox-core/hatchet"
	"github.com/nanobox-core/mist"
)

// structs
type (

	//
	Worker struct {
		sync.Mutex
		sync.WaitGroup

		queue []Job
		mist 	*mist.Mist
		log   hatchet.Logger
	}
)

// interfaces
type (

	//
	Job interface {
		Start(chan<- string)
		Collection() string
	}
)

//
func New(mist *mist.Mist, logger hatchet.Logger) *Worker {

	//
  if logger == nil {
    logger = hatchet.DevNullLogger{}
  }

  //
	if mist == nil {
		logger.Error("bonk")
	}

	worker := &Worker{
		queue: []Job{},
		mist:  mist,
		log: 	 logger,
	}

	return worker
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
	fmt.Println("Proccessing...")

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
			// log.Println("work failed:", err)
		}
	}()

	mist := make(chan string)

	go func() {
		job.Start(mist)
		close(mist)
	}()

	for data := range mist {
		w.mist.Publish([]string{job.Collection()}, data)
	}

}
