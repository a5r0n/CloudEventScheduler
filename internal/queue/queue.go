package queue

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/a5r0n/cloudeventscheduler/internal/cmd"
	"github.com/hibiken/asynq"
)

type Queue struct {
	Name      string
	client    *asynq.Client
	worker    *asynq.Server
	inspector *asynq.Inspector
}

func NewQueue(name string) *Queue {
	client, inspector := cmd.SetupAsynq()
	return &Queue{
		Name:      name,
		client:    client,
		inspector: inspector,
	}
}

func (q *Queue) NewTask(v interface{}) (*asynq.TaskInfo, error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload to json: %v", err)
	}
	task := asynq.NewTask(q.Name, payload)
	task_info, err := q.client.Enqueue(task)
	return task_info, err
}

func (q *Queue) GetInspector() *asynq.Inspector {
	if q.inspector == nil {
		log.Fatalln("inspector is nil")
	}
	return q.inspector
}
