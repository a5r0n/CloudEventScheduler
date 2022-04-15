package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/a5r0n/cloudeventscheduler/internal/cmd"
	"github.com/a5r0n/cloudeventscheduler/internal/task"
	"github.com/araddon/dateparse"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
	"github.com/spf13/viper"
)

var (
	client    *asynq.Client
	inspector *asynq.Inspector
	validate  = validator.New()
)

func enqueue(ctx context.Context, typename string, t task.WebhookTaskPayload, opts ...asynq.Option) (info *asynq.TaskInfo, err error) {
	if err := validate.Struct(t); err != nil {
		return nil, err
	}

	log.Printf(" [*] Send Webhook event to %s %s", typename, t.Url)
	payload, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	task := asynq.NewTask(typename, payload)
	info, err = client.EnqueueContext(ctx, task, opts...)
	if err == nil {
		log.Printf(" [*] Successfully enqueued task: %+v", info.ID)
	}
	return
}

func handleTaskPayload(c *fiber.Ctx) error {
	var t task.WebhookTaskPayload
	typename := c.Params("typename")

	queue := c.Params("queue")
	at := c.Query("at")
	dealy := c.Query("delay")

	var opts []asynq.Option

	if at != "" && dealy != "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "at and delay query params are mutually exclusive",
		})
	} else if at != "" {
		t, err := dateparse.ParseStrict(at)
		if err != nil {
			return err
		}
		opts = append(opts, asynq.ProcessAt(t))
	} else if dealy != "" {
		d, err := time.ParseDuration(dealy)
		if err != nil {
			return err
		}
		opts = append(opts, asynq.ProcessIn(d))
	}

	if queue != "" {
		opts = append(opts, asynq.Queue(queue))
	}

	if c.BodyParser(&t) != nil {
		return c.SendStatus(400)
	}

	if info, err := enqueue(c.Context(), typename, t, opts...); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error(), "success": false})
	} else {
		return c.Status(200).JSON(fiber.Map{"success": true, "task": info})
	}
}

func handleGetTaskInfo(c *fiber.Ctx) error {
	queue := c.Params("queue")
	taskid := c.Params("taskid")
	log.Printf(" [*] Get task info for %s %s", queue, taskid)
	info, err := inspector.GetTaskInfo(queue, taskid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error(), "success": false})
	} else {
		return c.Status(200).JSON(fiber.Map{"success": true, "task": info})
	}
}

func handleGetQueue(c *fiber.Ctx) error {
	queue := c.Params("queue")
	queueInfo, err := inspector.GetQueueInfo(queue)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error(), "success": false})
	}

	return c.Status(200).JSON(fiber.Map{"success": true, "queue": queueInfo})
}

func main() {
	cmd.SetupViper()
	cmd.EnsureRedisConfigs()

	client, inspector = cmd.SetupAsynq()
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Post("/:queue/:typename", handleTaskPayload)

	app.Get("/:queue/", handleGetQueue)
	app.Get("/:queue/:taskid", handleGetTaskInfo)

	cmd.ServeFiberApp(app, viper.GetInt("server.port"))

	log.Println("Running cleanup tasks...")

	if err := client.Close(); err != nil {
		log.Println(fmt.Printf("Error closing redis client: %s", err))
	}
	if err := inspector.Close(); err != nil {
		log.Println(fmt.Printf("Error closing redis inspector: %s", err))
	}
	log.Println("Bye!")
}
