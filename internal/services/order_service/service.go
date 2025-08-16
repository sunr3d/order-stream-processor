package order_service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/sunr3d/order-stream-processor/internal/interfaces/infra"
	"github.com/sunr3d/order-stream-processor/internal/interfaces/services"
	"github.com/sunr3d/order-stream-processor/models"
)

var _ services.OrderService = (*orderService)(nil)

type orderService struct {
	repo   infra.Database
	cache  infra.Cache
	broker infra.Broker
	logger *zap.Logger
}

func New(repo infra.Database, cache infra.Cache, broker infra.Broker, logger *zap.Logger) services.OrderService {
	return &orderService{
		repo:   repo,
		cache:  cache,
		broker: broker,
		logger: logger,
	}
}

func (s *orderService) ProcessOrder(ctx context.Context, order *models.Order) error {
	logger := s.logger.With(
		zap.String("op", "order_service.ProcessOrder"),
		zap.String("order_uid", order.OrderUID),
	)

	logger.Info("начинаем обработку заказа")

	// Сохранение заказа в БД
	if err := s.repo.Create(ctx, order); err != nil {
		if strings.Contains(err.Error(), "заказ уже существует") {
			logger.Info("заказ уже существует в БД")
			return err
		}
		logger.Error("ошибка при сохранении заказа в базе данных", zap.Error(err))
		return fmt.Errorf("repo.Create: %w", err)
	}

	// Сохранение заказа в кэш
	if err := s.cache.Set(ctx, order.OrderUID, order); err != nil {
		logger.Warn("ошибка при сохранении заказа в кэше", zap.Error(err))
	}

	logger.Info("заказ успешно обработан")
	return nil
}

func (s *orderService) GetOrder(ctx context.Context, orderUID string) (*models.Order, error) {
	logger := s.logger.With(
		zap.String("op", "order_service.GetOrder"),
		zap.String("order_uid", orderUID),
	)

	logger.Info("поиск заказа")

	// Поиск заказа в кэше
	order, err := s.cache.Get(ctx, orderUID)
	if err == nil {
		logger.Info("заказ был успешно найден в кэше")
		return order, nil
	}
	logger.Info("заказ не был найден в кэше, производим поиск в БД")

	// Поиск заказа в БД
	order, err = s.repo.Read(ctx, orderUID)
	if err != nil {
		logger.Error("ошибка при чтении заказа из базы данных", zap.Error(err))
		return nil, fmt.Errorf("repo.Read: %w", err)
	}

	// Сохранение заказа в кэш (т.к. ранее не был найден)
	if err := s.cache.Set(ctx, orderUID, order); err != nil {
		logger.Warn("ошибка при сохранении заказа в кэше", zap.Error(err))
	}

	logger.Info("заказ был успешно найден в БД")
	return order, nil
}

func (s *orderService) GetAllOrders(ctx context.Context) ([]*models.Order, error) {
	logger := s.logger.With(
		zap.String("op", "order_service.GetAllOrders"),
	)

	logger.Info("получение всех записей заказов из БД")

	// Получение всех записей заказов из БД
	orders, err := s.repo.ReadAll(ctx)
	if err != nil {
		logger.Error("ошибка при получении всех записей заказов из БД", zap.Error(err))
		return nil, fmt.Errorf("repo.ReadAll: %w", err)
	}

	// Сохранение всех записей заказов в кэш (или перезапись, если они уже существуют)
	if err := s.cache.Restore(ctx, orders); err != nil {
		logger.Warn("не удалось восстановить кэш", zap.Error(err))
	}

	logger.Info("все заказы успешно получены из БД",
		zap.Int("count", len(orders)),
	)
	return orders, nil
}

func (s *orderService) StartConsumer(ctx context.Context) error {
	s.broker.SetHandler(s.handleKafkaMessage)
	return s.broker.StartConsumer(ctx)
}

func (s *orderService) handleKafkaMessage(ctx context.Context, message []byte) error {
	logger := s.logger.With(
		zap.String("op", "order_service.handleKafkaMessage"),
	)

	var order models.Order
	if err := json.Unmarshal(message, &order); err != nil {
		logger.Error("ошибка при разборе сообщения из Kafka", zap.Error(err))
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	if err := s.ProcessOrder(ctx, &order); err != nil {
		logger.Error("ошибка при обработке заказа", zap.Error(err))
		return fmt.Errorf("order_service.ProcessOrder: %w", err)
	}

	logger.Info("заказ успешно обработан")
	return nil
}
