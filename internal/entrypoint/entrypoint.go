package entrypoint

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/sunr3d/order-stream-processor/internal/config"
	http_handlers "github.com/sunr3d/order-stream-processor/internal/handlers/http"
	kafka_handlers "github.com/sunr3d/order-stream-processor/internal/handlers/kafka"
	"github.com/sunr3d/order-stream-processor/internal/infra/inmem"
	"github.com/sunr3d/order-stream-processor/internal/infra/kafka"
	"github.com/sunr3d/order-stream-processor/internal/infra/postgres"
	"github.com/sunr3d/order-stream-processor/internal/interfaces/infra"
	"github.com/sunr3d/order-stream-processor/internal/middleware"
	"github.com/sunr3d/order-stream-processor/internal/server"
	"github.com/sunr3d/order-stream-processor/internal/services/order_service"
)

func Run(cfg *config.Config, logger *zap.Logger) error {
	logger.Info("запуск приложения...")

	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	/// Инфра слой
	db, err := postgres.New(cfg.Postgres, logger)
	if err != nil {
		logger.Error("ошибка при подключении к БД", zap.Error(err))
		return fmt.Errorf("postgres.New(): %w", err)
	}
	defer func(db infra.Database) {
		if closer, ok := db.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				logger.Error("ошибка при закрытии соединения с БД", zap.Error(err))
			} else {
				logger.Info("соединение с БД закрыто")
			}
		}
	}(db)

	cache := inmem.New(logger)

	broker, err := kafka.New(cfg.Kafka, logger)
	if err != nil {
		logger.Error("ошибка при подключении к Kafka", zap.Error(err))
		return fmt.Errorf("kafka.New(): %w", err)
	}
	defer func(broker infra.Broker) {
		if stopper, ok := broker.(interface{ Stop() error }); ok {
			if err := stopper.Stop(); err != nil {
				logger.Error("ошибка при закрытии соединения с Kafka", zap.Error(err))
			} else {
				logger.Info("соединение с Kafka закрыто")
			}
		}
	}(broker)

	/// Сервисный слой
	svc := order_service.New(db, cache, logger)

	/// HTTP слой
	controller := http_handlers.New(svc, logger)
	mux := http.NewServeMux()
	controller.RegisterOrderHandlers(mux)

	// Middleware
	handler := middleware.Recovery(logger)(
		middleware.ReqLogger(logger)(
			middleware.JSONValidator(logger)(mux),
		),
	)

	/// Kafka консьюмер
	consumerHandler := kafka_handlers.New(svc, logger)

	go func() {
		if err := broker.StartConsumer(appCtx, consumerHandler.CreateOrder); err != nil {
			logger.Error("ошибка при запуске консьюмера Kafka", zap.Error(err))
		}
	}()

	/// HTTP сервер
	srv := server.New(cfg.HTTPPort, handler, cfg.HTTPTimeout, logger)

	return srv.Start(appCtx)
}
