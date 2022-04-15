package main

import (
	"github.com/a5r0n/cloudeventscheduler/internal/cmd"
	"github.com/a5r0n/cloudeventscheduler/internal/queue"
	"github.com/hibiken/asynq"
)

func main() {
	cmd.SetupViper()
	q := queue.NewQueue("webhook:")
	q.SetupWorker(asynq.Config{Concurrency: 10})
	q.RunWorker()
}
