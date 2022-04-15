package provider

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/a5r0n/cloudeventscheduler/internal/task"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/spf13/viper"
	"github.com/xo/dburl"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	Grm *gorm.DB
)

type ScheduleEvent struct {
	gorm.Model
	ID        uuid.UUID               `gorm:"type:uuid;primary_key;"`
	TypeName  string                  `json:"typename" validate:"required"`
	QueueName string                  `json:"queuename" validate:"required"`
	Payload   task.WebhookTaskPayload `json:"payload" gorm:"embedded" validate:"required"`
	Cronspec  string                  `json:"cronspec" validate:"required"`
	TimeZone  string                  `json:"timezone"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (s *ScheduleEvent) BeforeCreate(tx *gorm.DB) error {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	uuid_str := uuid.String()
	tx.Statement.SetColumn("ID", uuid_str)
	return nil
}

func (s *ScheduleEvent) GetTask() (*asynq.Task, error) {
	if s.TypeName == "" {
		return nil, errors.New("typename is required")
	}
	payload, err := json.Marshal(s.Payload)
	if err != nil {
		return nil, err
	}
	task := asynq.NewTask(s.TypeName, payload)
	return task, nil
}

func Setup() error {
	var err error
	u, err := dburl.Parse(viper.GetString("database.dsn"))
	if err != nil {
		log.Fatalln(err)
	}

	switch u.Driver {
	case "sqlite3":
		log.Printf("Using sqlite3 %s\n", u.DSN)
		Grm, err = gorm.Open(sqlite.Open(u.DSN), &gorm.Config{})
	case "postgres":
		log.Printf("Using postgres %s\n", u.DSN)
		Grm, err = gorm.Open(postgres.Open(u.DSN), &gorm.Config{})
	default:
		log.Fatalf("unsupported driver %s", u.Driver)
	}

	if err != nil {
		return err
	}
	Grm.AutoMigrate(&ScheduleEvent{})
	return err
}
