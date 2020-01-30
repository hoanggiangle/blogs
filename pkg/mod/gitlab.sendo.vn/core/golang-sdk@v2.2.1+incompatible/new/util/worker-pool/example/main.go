package main

import (
	"fmt"
	"gitlab.sendo.vn/core/golang-sdk/util/worker-pool"
	"log"
	"time"
)

func main() {
	pool, err := sdwp.NewWorkerPool(5, 100)
	if err != nil {
		log.Fatal(err)
	}

	// Show log of the pool
	pool.SetLogEnable(true)
	// Start all workers in pool
	pool.Start()

	// Add 200 tasks into pool
	for i := 1; i <= 200; i++ {
		task := sdwp.NewTask(fmt.Sprintf("%d", i), func() error {
			time.Sleep(500 * time.Millisecond)
			return nil
		})

		pool.AddTask(task)
	}

	// Stop pool after 5s
	time.Sleep(5 * time.Second)
	pool.Stop()

	// Restart pool after 4s
	time.Sleep(4 * time.Second)
	pool.Start()

	<-pool.Done()
	log.Println("Pool is cleared")
}
