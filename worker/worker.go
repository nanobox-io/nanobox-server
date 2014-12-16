package worker

import (
	"fmt"
	"sync"

	"github.com/nanobox-core/hatchet"
)

// structs
type (

	//
	Worker struct {
		sync.Mutex
		sync.WaitGroup

		queue []Job
		publisher 	Publisher
		log   hatchet.Logger
	}

// maybe?
	channelPublisher struct {
		ch chan string
	}
)

// interfaces
type (

	//
	Job interface {
		Start(chan<- string)
		Collection() string
	}

	Publisher interface {
		Publish(tags []string, data string)
	}

)

//
func New(publisher Publisher, logger hatchet.Logger) *Worker {

	//
  if logger == nil {
    logger = hatchet.DevNullLogger{}
  }

  //
	if publisher == nil {
		logger.Error("bonk")
	}

	worker := &Worker{
		queue: []Job{},
		publisher:  publisher,
		log: 	 logger,
	}

	return worker
}

// maybe?
func NewSub(ch chan string, logger hatchet.Logger) *Worker {
	cPub := channelPublisher{ch}
	return New(cPub, logger)
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

	publisher := make(chan string)

	go func() {
		job.Start(publisher)
		close(publisher)
	}()

	for data := range publisher {
		w.publisher.Publish([]string{job.Collection()}, data)
	}

}

func (c *channelPublisher) Publish(tags []string, data string) {
	c.ch <- fmt.Sprintf("%v %d", tags, data)
}

