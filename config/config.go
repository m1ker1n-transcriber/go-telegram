package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type Config struct {
	Telegram TelegramConfig `env-prefix:"TELEGRAM_"`
	Minio    MinioConfig    `env-prefix:"MINIO_"`
}

type TelegramConfig struct {
	ApiToken string `env:"API_TOKEN" env-required:"true"`
}

type MinioConfig struct {
	Endpoint      string        `env:"ENDPOINT" env-required:"true"`
	Region        string        `env:"REGION" env-default:"us-east-1"`
	AccessKey     string        `env:"ACCESS_KEY" env-required:"true"`
	SecretKey     string        `env:"SECRET_KEY" env-required:"true"`
	BucketName    string        `env:"BUCKET_NAME" env-required:"true"`
	UploadTimeout time.Duration `env:"UPLOAD_TIMEOUT" env-default:"10m"`
}

func MustLoad() Config {
	cfg := &Config{}

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		panic(fmt.Errorf("could not receive config: %w", err))
	}

	return *cfg
}
