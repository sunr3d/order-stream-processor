package kafka_handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/sunr3d/order-stream-processor/internal/handlers/validators"
	"github.com/sunr3d/order-stream-processor/internal/interfaces/services"
	"github.com/sunr3d/order-stream-processor/models"
)

type kafkaHandler struct {
	svc    services.OrderService
	logger *zap.Logger
}

func New(svc services.OrderService, logger *zap.Logger) *kafkaHandler {
	return &kafkaHandler{svc: svc, logger: logger}
}

func (h *kafkaHandler) CreateOrder(ctx context.Context, msg []byte) error {
	logger := h.logger.With(zap.String("op", "kafka_handlers.createOrder"))

	var order models.Order
	if err := json.Unmarshal(msg, &order); err != nil {
		logger.Error("ошибка при разборе заказа из Kafka",
			zap.Error(err),
			zap.String("order_uid", order.OrderUID),
		)
		return fmt.Errorf("ошибка при разборе заказа из Kafka: %w", err)
	}

	if err := validators.ValidateOrder(&order); err != nil {
		logger.Error("ошибка валидации заказа из Kafka",
			zap.Error(err),
		)
		return fmt.Errorf("ошибка валидации заказа из Kafka: %w", err)
	}

	logger = logger.With(zap.String("order_uid", order.OrderUID))

	if err := h.svc.ProcessOrder(ctx, &order); err != nil {
		logger.Error("ошибка при обработке заказа из Kafka",
			zap.Error(err),
			zap.String("order_uid", order.OrderUID),
		)
		return fmt.Errorf("order_service.ProcessOrder(): %w", err)
	}

	logger.Info("заказ из Kafka успешно обработан")
	return nil
}
