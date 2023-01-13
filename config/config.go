package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Port        int    `env:"PORT" envDefault:"3000"`
	DatabaseUrl string `env:"DATABASE_URL" envDefault:"3000"`
}

func Parse() Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("%+v\n", err)
	}
	return cfg
}
