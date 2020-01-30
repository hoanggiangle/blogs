package sdwp

import (
	"errors"
	"sync"
)

type WorkerPool struct {
	m              *sync.RWMutex
	poolChan       chan Task
	workers        []Worker
	maxConcurrent  int
	stopChan       chan bool // use for stopping all workers
	finishedChan   chan bool // use for listening when all task is finished
	workerDoneTask chan bool // use for listening when worker has done a task
	isLogEnabled   bool
	isCleared      bool
}

// Create a new worker pool
//   * maxCon int: Maximum concurrent is allowed in pool (must be > 0)
//   * maxWaiter int: Maximum tasks in queue (must be > 0)
func NewWorkerPool(maxCon, maxWaiter int) (WorkerPool, error) {
	if maxCon < 1 || maxWaiter < 1 {
		return WorkerPool{}, errors.New("max concurrent and max waiter can not less than 1")
	}

	return WorkerPool{
		m:              &sync.RWMutex{},
		poolChan:       make(chan Task, maxWaiter),
		maxConcurrent:  maxCon,
		stopChan:       make(chan bool),
		workerDoneTask: make(chan bool),
		workers:        []Worker{},
		isLogEnabled:   false,
		isCleared:      false,
	}, nil
}

func (p *WorkerPool) SetLogEnable(isEnabled bool) {
	p.isLogEnabled = isEnabled

	for _, w := range p.workers {
		w.isLogEnabled = isEnabled
	}
}

func (p *WorkerPool) Stop() {
	for _, w := range p.workers {
		w.stopChan <- true
	}
	p.workers = []Worker{}
	p.stopChan <- true
}

func (p *WorkerPool) Start() {
	for i := 1; i <= p.maxConcurrent; i++ {
		w := newWorker(uint(i), p.isLogEnabled)
		p.workers = append(p.workers, w)
		go w.start(p.poolChan, p.workerDoneTask)
	}

	go func() {
		for {
			select {
			case <-p.workerDoneTask:
				remain := len(p.poolChan)
				if remain == 0 && p.finishedChan != nil {
					// Fixed: Data racing on p.isCleared
					// cause deadlock on p.finishedChan
					p.m.Lock()
					if !p.isCleared {
						p.isCleared = true
						// Ensure only fire finishedChan once
						p.finishedChan <- true
					}
					p.m.Unlock()
				}

				// if p.isLogEnabled {
				// 	log.Printf("Remaining %d tasks", remain)
				// }
			case <-p.stopChan:
				return
			}
		}
	}()
}

func (p *WorkerPool) AddTask(task Task) {
	p.isCleared = false
	go func() {
		p.poolChan <- task
	}()
}

func (p *WorkerPool) Done() <-chan bool {
	fc := make(chan bool)
	p.finishedChan = fc
	return fc
}
