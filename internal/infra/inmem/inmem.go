package inmem

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/sunr3d/order-stream-processor/internal/interfaces/infra"
	"github.com/sunr3d/order-stream-processor/models"
)

var _ infra.Cache = (*inmemCache)(nil)

type inmemCache struct {
	data   map[string]*models.Order
	mu     sync.RWMutex
	logger *zap.Logger
}

func New(log *zap.Logger) infra.Cache {
	return &inmemCache{
		data:   make(map[string]*models.Order),
		logger: log,
	}
}

func (c *inmemCache) Set(ctx context.Context, orderUID string, order *models.Order) error {
	logger := c.logger.With(
		zap.String("op", "inmem.Set"),
		zap.String("order_uid", orderUID),
	)

	logger.Info("сохранение заказа в кэше...")

	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[orderUID] = order

	logger.Info("заказ успешно сохранен в кэше")
	return nil
}

func (c *inmemCache) Get(ctx context.Context, orderUID string) (*models.Order, error) {
	logger := c.logger.With(
		zap.String("op", "inmem.Get"),
		zap.String("order_uid", orderUID),
	)

	logger.Info("поиск заказа в кэше...")

	c.mu.RLock()
	defer c.mu.RUnlock()

	order, exists := c.data[orderUID]
	if !exists {
		logger.Info("заказ не найден в кэше")
		return nil, fmt.Errorf("заказ не найден в кэше: %s", orderUID)
	}

	logger.Info("заказ успешно найден в кэше")
	return order, nil
}

func (c *inmemCache) Restore(ctx context.Context, orders []*models.Order) error {
	logger := c.logger.With(
		zap.String("op", "inmem.Restore"),
		zap.Int("count", len(orders)),
	)

	logger.Info("восстановление заказов в кэш...")

	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*models.Order)
	for _, order := range orders {
		c.data[order.OrderUID] = order
	}

	logger.Info("все заказы успешно восстановлены", zap.Int("restored_count", len(c.data)))

	return nil
}
