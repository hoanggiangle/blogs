package sdwp

import "log"

type Worker struct {
	id           uint
	stopChan     chan bool
	isLogEnabled bool
}

func (w *Worker) start(input <-chan Task, done chan bool) {
	go func() {
		for {
			select {
			case task := <-input:
				err := task.handler()
				if w.isLogEnabled {
					log.Printf("Worker %d done task %s with error: %v \n", w.id, task.GetId(), err)
				}
				done <- true
			case <-w.stopChan:
				if w.isLogEnabled {
					log.Printf("Worker %d is stopped", w.id)
				}
				return
			}
		}
	}()
}

func newWorker(id uint, isLogEnabled bool) Worker {
	return Worker{
		id:           id,
		stopChan:     make(chan bool),
		isLogEnabled: isLogEnabled,
	}
}
