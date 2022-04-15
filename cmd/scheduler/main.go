package main

import (
	"log"
	"time"

	"github.com/a5r0n/cloudeventscheduler/internal/cmd"
	"github.com/a5r0n/cloudeventscheduler/pkg/provider"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
	"github.com/spf13/viper"
)

var (
	validate = validator.New()
)

func handleNewSchudleEvent(c *fiber.Ctx) error {
	var event provider.ScheduleEvent
	if err := c.BodyParser(&event); err != nil {
		return err
	}

	if err := validate.Struct(event); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	if err := provider.Grm.Save(&event).Error; err != nil {
		return err
	}

	return c.Status(201).JSON(&event)
}

func handleDeleteSchudleEvent(c *fiber.Ctx) error {
	eventId := c.Params("id")
	result := provider.Grm.Delete(&provider.ScheduleEvent{}, "id = ?", eventId)

	if result.Error != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   result.Error.Error(),
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"error":   "event not found",
		})
	}

	return c.SendStatus(204)
}

func handleGetAllSchudleEvents(c *fiber.Ctx) error {
	var events []provider.ScheduleEvent
	if err := provider.Grm.Find(&events).Error; err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"events":  events,
	})
}

func main() {
	cmd.SetupViper()
	cmd.EnsureRedisConfigs()
	cmd.EnsureDatabaseConfigs()

	err := provider.Setup()
	if err != nil {
		log.Fatal(err)
	}
	provider := provider.NewDatabaseBasedConfigProvider()

	mgr, err := asynq.NewPeriodicTaskManager(
		asynq.PeriodicTaskManagerOpts{
			RedisConnOpt:               cmd.GetRedisOpts(),
			PeriodicTaskConfigProvider: provider,        // this provider object is the interface to your config source
			SyncInterval:               2 * time.Second, // this field specifies how often sync should happen
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	if err := mgr.Start(); err != nil {
		log.Fatal(err)
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/", handleGetAllSchudleEvents)
	app.Post("/", handleNewSchudleEvent)
	app.Delete("/:id", handleDeleteSchudleEvent)

	cmd.ServeFiberApp(app, viper.GetInt("server.port"))
	log.Println("Running cleanup tasks...")
	mgr.Shutdown()
}
