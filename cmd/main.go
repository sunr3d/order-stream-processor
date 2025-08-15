package main

import (
	"log"

	"github.com/sunr3d/order-stream-processor/internal/config"
	"github.com/sunr3d/order-stream-processor/internal/logger"
	"github.com/sunr3d/order-stream-processor/internal/entrypoint"
)

func main() {
	cfg, err := config.GetConfigFromEnv()
	if err != nil {
		log.Fatalf("ошибка при загрузке конфигурации: %s\n", err.Error())
	}

	zapLogger := logger.New(cfg.LogLevel)

	if err = entrypoint.Run(cfg, zapLogger); err != nil {
		log.Fatalf("ошибка при запуске приложения: %s\n", err.Error())
	}
}
