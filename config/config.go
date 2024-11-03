// config.go
package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	SessionSecret      string `mapstructure:"SESSION_SECRET"`
	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	RedirectURL        string `mapstructure:"GOOGLE_LOGIN_REDIRECT_URL"`
	MongoURI           string `mapstructure:"MONGO_URI"`
	MongoDatabase      string `mapstructure:"MONGO_DATABASE"`
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

	// Validate required environment variables
	if config.SessionSecret == "" {
		return nil, fmt.Errorf("SESSION_SECRET environment variable is required")
	}
	if config.GoogleClientID == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID environment variable is required")
	}
	if config.GoogleClientSecret == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_SECRET environment variable is required")
	}
	if config.RedirectURL == "" {
		return nil, fmt.Errorf("GOOGLE_LOGIN_REDIRECT_URL environment variable is required")
	}

	return &config, nil
}
