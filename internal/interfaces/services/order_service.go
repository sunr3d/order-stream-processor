package services

import (
	"context"

	"github.com/sunr3d/order-stream-processor/models"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.2 --name=OrderService --output=../../../mocks --filename=mock_order_service.go --with-expecter
type OrderService interface {
	ProcessOrder(ctx context.Context, data []byte) error
	GetOrder(ctx context.Context, orderUID string) (*models.Order, error)
	GetAllOrders(ctx context.Context) ([]*models.Order, error)
}
