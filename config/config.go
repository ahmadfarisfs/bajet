// config.go
package config

import (
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	SessionSecret        string `mapstructure:"SESSION_SECRET"`
	SQLitePath           string `mapstructure:"SQLITE_PATH"`
	DefaultBudget        int    `mapstructure:"DEFAULT_BUDGET" validate:"gt=0"`
	CycleStartDay        int    `mapstructure:"CYCLE_START_DAY" validate:"gte=1,lte=28"`
	ContextTimeoutSecond int    `mapstructure:"CONTEXT_TIMEOUT_SECOND" validate:"gt=0"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Println("No .env file found, using machine environment variables")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if config.SQLitePath == "" {
		config.SQLitePath = "bajet.db"
	}
	if config.SessionSecret == "" {
		config.SessionSecret = "bajet-dev-session-secret"
	}
	if config.DefaultBudget == 0 {
		config.DefaultBudget = 2_000_000
	}
	if config.CycleStartDay == 0 {
		config.CycleStartDay = 25
	}
	if config.ContextTimeoutSecond == 0 {
		config.ContextTimeoutSecond = 10
	}

	// Validate the config using the validator package
	validate := validator.New()
	if err := validate.Struct(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %v", err)
	}

	return &config, nil
}
