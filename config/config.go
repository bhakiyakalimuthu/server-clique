package config

import (
	"github.com/caarlos0/env"
	"github.com/go-playground/validator/v10"
)

// Config contains environment values
type Config struct {
	DebugLog bool `env:"DEBUG_LOG" envDefault:"false"`
	LogJSON  bool `env:"LOG_JSON" envDefault:"true"`
	// AppName  string `env:"APP_NAME" envDefault:"server-clique"`
	// Connection string of message queue (rabbitmq is used)
	QueueConnString string `env:"QUEUE_CONN_STRING" envDefault:"amqp://guest:guest@localhost:5672/"`
	// Queue name
	QueueName      string `env:"QUEUE_NAME" envDefault:"rabbit-queue"`
	OutputFileName string `env:"OUTPUT_FILE_NAME" envDefault:"output.json"`
}

// LoadFromEnv parses environment variables into a given struct and validates
// its fields' values.
func LoadFromEnv(config interface{}) error {
	if err := env.Parse(config); err != nil {
		return err
	}
	if err := validator.New().Struct(config); err != nil {
		return err
	}
	return nil
}

func NewConfig() *Config {
	var cfg Config
	if err := LoadFromEnv(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}
