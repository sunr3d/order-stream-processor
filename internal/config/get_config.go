package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

func GetConfigFromEnv() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("Не удалось загрузить .env файл: \"%s\", продолжаем со значениями окружения по умолчанию\n", err.Error())
	}
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("envconfig.Process: %w", err)
	}

	if brokers := cfg.Kafka.Brokers; len(brokers) == 1 && strings.Contains(brokers[0], ",") {
		cfg.Kafka.Brokers = strings.Split(brokers[0], ",")
	}
	return cfg, nil
}
