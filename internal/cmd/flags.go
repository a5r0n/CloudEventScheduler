package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func SetupAsynq() (client *asynq.Client, inspector *asynq.Inspector) {
	redisOpts := GetRedisOpts()
	log.Printf(" [*] Connecting to redis at %s/%d", redisOpts.Addr, redisOpts.DB)
	client = asynq.NewClient(redisOpts)
	inspector = asynq.NewInspector(redisOpts)
	return
}

func GetRedisOpts() asynq.RedisClientOpt {

	redisUri := viper.GetString("redis")
	redisConn, err := asynq.ParseRedisURI(redisUri)
	if err != nil {
		log.Fatal(fmt.Errorf("error parsing redis uri: %s", err))
	}

	r, ok := redisConn.(asynq.RedisClientOpt)

	if !ok {
		log.Fatal(fmt.Errorf("error parsing redis uri: %s", err))
	}

	return r
}

// SetupViper is a helper function to setup viper with config paths, env, and flags.
func SetupViper() {
	setupPFlags()
	viper.SetEnvPrefix("asynq")
	viper.AutomaticEnv()

	viper.SetConfigName(".config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/asynqevents")
	viper.AddConfigPath("$HOME/.config/asynqevents")
	viper.AddConfigPath("$HOME/.asynqevents")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file, %s\n", err)
	}
}

func setupPFlags() {
	// setup server flags
	flag.Int("server.port", 3000, "port for the server to listen on")

	// setup redis flags
	flag.String("redis", "", "redis uri")

	// setup database flags
	flag.String("database.dsn", "", "database dsn")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}

func EnsureRedisConfigs() {
	if viper.GetString("redis") == "" {
		log.Fatal("redis uri is required")
	}
}

func EnsureServerConfigs() {
	if viper.GetString("server.port") == "" {
		log.Fatal("server port is required")
	}
}

func EnsureDatabaseConfigs() {
	if viper.GetString("database.dsn") != "" {
		return
	}
	if viper.GetString("database.host") == "" {
		log.Fatal("database host is required")
	}
	if viper.GetString("database.port") == "" {
		log.Fatal("database port is required")
	}
	if viper.GetString("database.user") == "" {
		log.Fatal("database user is required")
	}
	if viper.GetString("database.password") == "" {
		log.Fatal("database password is required")
	}
	if viper.GetString("database.name") == "" {
		log.Fatal("database name is required")
	}
}

// ServeFiberApp is a helper function to start a fiber app. and gracefully shutdown it.
func ServeFiberApp(app *fiber.App, port int) error {
	log.Printf(" [*] Listening on port :%d", port)
	if err := GracefullyShutdown(
		func() error { return app.Listen(fmt.Sprintf(":%d", port)) },
		func() error { return app.Shutdown() },
	); err != nil {
		log.Println(err)
	}
	return nil
}

// gracefullyShutdown is a helper function to gracefully shutdown a server.
func GracefullyShutdown(start func() error, stop func() error) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	serverShutdown := make(chan struct{})

	go func() {
		<-c
		log.Println("Gracefully shutting down...")
		_ = stop()
		serverShutdown <- struct{}{}
	}()

	if err := start(); err != nil {
		log.Println(err)
	}
	<-serverShutdown
	return nil
}
