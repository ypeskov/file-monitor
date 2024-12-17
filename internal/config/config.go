package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Port           string `env:"SERVER_PORT" envDefault:"80"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
	LogAllRequests bool   `env:"LOG_ALL_REQUESTS" envDefault:"false"`
}

func New() *Config {
	cfg := &Config{}
	if err := godotenv.Load(); err != nil {
		log.Warn("No .env file found")
	}
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}

	cfg.LogLevel = "info"

	return cfg
}
