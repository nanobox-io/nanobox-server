package worker_test

import "testing"
import "github.com/nanobox-io/nanobox-server/util"

type Job struct {
	ProcessCount int
}

func (j *Job) Process() {
	j.ProcessCount++
}

func TestWorkerQueue(t *testing.T) {
	w := worker.New()
	w.Blocking = true
	j := &Job{0}
	w.Queue(j)
	if w.Count() != 1 {
		t.Errorf("There should be 1 job queued in the worker but there is %d", w.Count())
	}
	w.Process()
	if w.Count() != 0 {
		t.Errorf("There should be 0 job queued in the worker but there is %d", w.Count())
	}
	if j.ProcessCount != 1 {
		t.Errorf("The Job should have been processed 1 time")
	}
}

func TestWorkerQueueAndProcess(t *testing.T) {
	w := worker.New()
	w.Blocking = true
	j := &Job{0}
	w.QueueAndProcess(j)
	if w.Count() != 0 {
		t.Errorf("There should be 1 job queued in the worker but there is %d", w.Count())
	}
	if j.ProcessCount != 1 {
		t.Errorf("The Job should have been processed 1 time")
	}

}
