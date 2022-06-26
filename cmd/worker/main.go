package main

import (
	"github.com/a5r0n/cloudeventscheduler/internal/cmd"
	"github.com/a5r0n/cloudeventscheduler/internal/queue"
	"github.com/hibiken/asynq"
)

var config cmd.AppConfig

func main() {
	config = cmd.SetupViper()
	q := queue.NewQueue("webhook:")
	q.SetupWorker(asynq.Config{Concurrency: 10})
	q.RunWorker()
}
