package order_service

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/sunr3d/order-stream-processor/internal/interfaces/infra"
	"github.com/sunr3d/order-stream-processor/internal/interfaces/services"
	"github.com/sunr3d/order-stream-processor/models"
)

var _ services.OrderService = (*orderService)(nil)

type orderService struct {
	repo   infra.Database
	cache  infra.Cache
	logger *zap.Logger
}

func New(repo infra.Database, cache infra.Cache, logger *zap.Logger) services.OrderService {
	return &orderService{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

func (s *orderService) ProcessOrder(ctx context.Context, data []byte) error {
	logger := s.logger.With(
		zap.String("op", "order_service.ProcessOrder"),
	)

	// Парсинг JSON
	var order models.Order
	if err := json.Unmarshal(data, &order); err != nil {
		logger.Error("ошибка при парсинге заказа", zap.Error(err))
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	logger = logger.With(zap.String("order_uid", order.OrderUID))
	logger.Info("обработка заказа")

	// Валидация заказа
	if err := s.validateOrderJSON(&order); err != nil {
		logger.Error("ошибка при валидации заказа", zap.Error(err))
		return fmt.Errorf("validateOrder: %w", err)
	}

	// Проверка на дубликат в БД
	exists, err := s.repo.Read(ctx, order.OrderUID)
	if err != nil {
		logger.Error("ошибка при чтении заказа из базы данных", zap.Error(err))
		return fmt.Errorf("repo.Read: %w", err)
	}
	if exists != nil {
		logger.Info("заказ уже существует в базе данных")
		return fmt.Errorf("заказ уже существует в базе данных")
	}

	// Сохранение заказа в БД
	if err := s.repo.Create(ctx, &order); err != nil {
		logger.Error("ошибка при сохранении заказа в базе данных", zap.Error(err))
		return fmt.Errorf("repo.Create: %w", err)
	}

	// Сохранение заказа в кэш
	if err := s.cache.Set(ctx, order.OrderUID, &order); err != nil {
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

// Вспомогательные функции
// Валидация заказа (только основные поля, можно расширить по мере необходимости)
func (s *orderService) validateOrderJSON(order *models.Order) error {
	// Основные поля
	if order.OrderUID == "" {
		return fmt.Errorf("order_uid не может быть пустым")
	}
	if order.CustomerID == "" {
		return fmt.Errorf("customer_id не может быть пустым")
	}
	if len(order.Items) == 0 {
		return fmt.Errorf("items не может быть пустым")
	}
	if order.TrackNumber == "" {
		return fmt.Errorf("track_number не может быть пустым")
	}

	// Поля доставки
	if order.Delivery.Name == "" {
		return fmt.Errorf("delivery.name не может быть пустым")
	}

	// Поля платежа
	if order.Payment.Transaction == "" {
		return fmt.Errorf("payment.transaction не может быть пустым")
	}
	if order.Payment.Provider == "" {
		return fmt.Errorf("payment.provider не может быть пустым")
	}
	if order.Payment.Amount <= 0 {
		return fmt.Errorf("payment.amount не может быть меньше или равно 0")
	}
	if order.Payment.PaymentDT <= 0 {
		return fmt.Errorf("payment.payment_dt не может быть меньше или равно 0")
	}

	return nil
}
