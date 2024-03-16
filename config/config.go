package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	TelegramApiToken string `env:"TELEGRAM_API_TOKEN" env-required:"true"`
}

func MustLoad() Config {
	cfg := &Config{}

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		panic(fmt.Errorf("could not receive config: %w", err))
	}

	return *cfg
}
