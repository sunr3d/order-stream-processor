package infra

import (
	"context"

	"github.com/sunr3d/order-stream-processor/models"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.2 --name=Cache --output=../../../mocks --filename=mock_cache.go --with-expecter
type Cache interface {
	Set(ctx context.Context, orderUID string, order *models.Order) error
	Get(ctx context.Context, orderUID string) (*models.Order, error)
	Restore(ctx context.Context, orders []*models.Order) error
}
