// config.go
package config

import (
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	SessionSecret        string `mapstructure:"SESSION_SECRET" validate:"required"`
	GoogleClientID       string `mapstructure:"GOOGLE_CLIENT_ID" validate:"required"`
	GoogleClientSecret   string `mapstructure:"GOOGLE_CLIENT_SECRET" validate:"required"`
	RedirectURL          string `mapstructure:"GOOGLE_LOGIN_REDIRECT_URL" validate:"required,url"`
	MongoURI             string `mapstructure:"MONGO_URI" validate:"required,url"`
	MongoDatabase        string `mapstructure:"MONGO_DATABASE" validate:"required"`
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

	// Validate the config using the validator package
	validate := validator.New()
	if err := validate.Struct(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %v", err)
	}

	return &config, nil
}
