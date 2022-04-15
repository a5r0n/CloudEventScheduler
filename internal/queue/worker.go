package queue

import (
	"log"

	"github.com/a5r0n/cloudeventscheduler/internal/cmd"
	"github.com/hibiken/asynq"
)

func (q *Queue) RunWorker() {
	log.Printf(" [*] Starting handler for %s", q.Name)

	log.Printf(" [*] Starting worker for %s", q.Name)
	cmd.GracefullyShutdown(
		func() error { return q.worker.Run(asynq.HandlerFunc(q.handle)) },
		func() error { return nil },
	)

	log.Printf(" [*] Stopping worker for %s", q.Name)
}

func (q *Queue) SetupWorker(config asynq.Config) {
	log.Printf(" [*] Setting up worker for %s", q.Name)
	q.worker = asynq.NewServer(
		cmd.GetRedisOpts(),
		config,
	)
}
