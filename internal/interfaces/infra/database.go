package infra

import (
	"context"

	"github.com/sunr3d/order-stream-processor/models"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.2 --name=Database --output=../../../mocks --filename=mock_database.go --with-expecter
type Database interface {
	Create(ctx context.Context, order *models.Order) error
	Read(ctx context.Context, orderUID string) (*models.Order, error)
	ReadAll(ctx context.Context) ([]*models.Order, error)
}
