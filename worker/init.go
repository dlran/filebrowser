package worker

import (
	"context"
	"github.com/hibiken/asynq"
	"log"
	"time"
)

func initHandler(ctx context.Context, task *asynq.Task) error {
	taskID, ok := asynq.GetTaskID(ctx)
	if !ok {
		log.Println("failed to get task id")
		return nil
	}
	retryCount, _ := asynq.GetRetryCount(ctx)
	maxRetry, _ := asynq.GetMaxRetry(ctx)
	log.Printf("Processing task %s ID:%s %d %d", task.Type(), taskID, retryCount, maxRetry)

	var res string
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		log.Printf("doing nothing x%d...", i)
	}
	res = "complete"

	if _, err := task.ResultWriter().Write([]byte(res)); err != nil {
		log.Printf("failed to write task result: %v", err)
		return err
	}

	log.Println("task done")
	return nil
}
