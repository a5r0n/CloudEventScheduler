package task

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
	"gorm.io/datatypes"
)

type TaskPayload interface {
}

type WebhookTaskPayload struct {
	Url     string         `json:"url" validate:"required"`
	Method  string         `json:"method" validate:"required"`
	Body    datatypes.JSON `json:"body"`
	Headers datatypes.JSON `json:"headers"`
}

func HandleWebhookTask(ctx context.Context, t *asynq.Task) error {
	var p WebhookTaskPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	log.Printf(" [*] Send Webhook event to %s", p.Url)
	return nil
}
