package provider

import (
	"context"
	"log"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// DatabaseBasedConfigProvider is a config provider that reads configs from database
// implements asynq.PeriodicTaskConfigProvider interface.
type DatabaseBasedConfigProvider struct {
	db *gorm.DB
}

func NewDatabaseBasedConfigProvider() *DatabaseBasedConfigProvider {
	return &DatabaseBasedConfigProvider{
		db: Grm,
	}
}

func (p *DatabaseBasedConfigProvider) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	var configs []*asynq.PeriodicTaskConfig
	var events []ScheduleEvent

	if err := p.db.Find(&events).Error; err != nil {
		return nil, err
	}
	for _, v := range events {
		task, err := v.GetTask()
		if err != nil {
			log.Printf("failed to get task: %v", err)
			continue
		}
		configs = append(configs, &asynq.PeriodicTaskConfig{
			Cronspec: v.Cronspec,
			Task:     task,
			Opts:     []asynq.Option{asynq.Queue(v.QueueName)},
		})
	}

	return configs, nil
}

func (p *DatabaseBasedConfigProvider) PingContext(ctx context.Context) error {
	db, err := p.db.DB()
	if err != nil {
		return err
	}
	return db.PingContext(ctx)
}
