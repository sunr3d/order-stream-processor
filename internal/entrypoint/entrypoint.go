package entrypoint

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/sunr3d/order-stream-processor/internal/api"
	"github.com/sunr3d/order-stream-processor/internal/config"
	"github.com/sunr3d/order-stream-processor/internal/infra/inmem"
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

	// Инфра слой
	db, err := postgres.New(cfg.Postgres, logger)
	if err != nil {
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

	// Сервисный слой
	svc := order_service.New(db, cache, logger)

	// API слой
	controller := api.New(svc, logger)
	mux := http.NewServeMux()
	controller.RegisterOrderHandlers(mux)

	// Middleware слой
	handler := middleware.Recovery(logger)(
		middleware.ReqLogger(logger)(
			middleware.JSONValidator(logger)(mux),
		),
	)

	// HTTP сервер
	srv := server.New(cfg.HTTPPort, handler, cfg.HTTPTimeout, logger)

	return srv.Start(appCtx)
}
