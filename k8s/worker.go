package k8s

import (
	"context"
	"log"
	"sync"
	"time"

	"jabberwocky238/storebirth/dblayer"
)

var (
	workerOnce   sync.Once
	workerCtx    context.Context
	workerCancel context.CancelFunc
)

// StartWorker starts the background task worker
func StartWorker() {
	workerOnce.Do(func() {
		workerCtx, workerCancel = context.WithCancel(context.Background())
		go taskWorker(workerCtx)
		log.Println("Config task worker started")
	})
}

// StopWorker stops the background task worker
func StopWorker() {
	if workerCancel != nil {
		workerCancel()
	}
}

// taskWorker processes pending config tasks
func taskWorker(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Config task worker stopped")
			return
		case <-ticker.C:
			processPendingTasks()
		}
	}
}

// processPendingTasks fetches and processes pending tasks
func processPendingTasks() {
	if dblayer.DB == nil {
		return
	}

	// Fetch one pending task (FIFO)
	taskID, userUID, taskType, err := dblayer.FetchPendingTask()
	if err != nil {
		return // No pending tasks
	}

	log.Printf("Processing task %d: %s for user %s", taskID, taskType, userUID)

	// Execute task
	var taskErr error
	switch taskType {
	case "config_update":
		taskErr = UpdateCombinatorConfig(userUID)
	case "pod_create":
		taskErr = CreateCombinatorPod(userUID)
	default:
		log.Printf("Unknown task type: %s", taskType)
	}

	// Update task status
	if taskErr != nil {
		dblayer.MarkTaskFailed(taskID, taskErr.Error())
		log.Printf("Task %d failed: %v", taskID, taskErr)
	} else {
		dblayer.MarkTaskCompleted(taskID)
		log.Printf("Task %d completed", taskID)
	}
}
