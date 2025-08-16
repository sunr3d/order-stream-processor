package order_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/sunr3d/order-stream-processor/internal/services/order_service"
	"github.com/sunr3d/order-stream-processor/mocks"
	"github.com/sunr3d/order-stream-processor/models"
)

func createValidOrder() *models.Order {
	return &models.Order{
		OrderUID:    "test-123",
		CustomerID:  "customer-123",
		TrackNumber: "TRACK-123",
		Items: []models.Item{
			{Name: "Test Item 1", Price: 100, TotalPrice: 100},
			{Name: "Test Item 2", Price: 200, TotalPrice: 200},
		},
		Delivery: models.Delivery{
			Name: "Test User",
		},
		Payment: models.Payment{
			Transaction: "transaction-123",
			Provider:    "test-provider",
			Amount:      300,
			PaymentDT:   time.Now().Unix(),
		},
	}
}

// ProcessOrder Tests
func TestOrderService_ProcessOrder_OK(t *testing.T) {
	repo := &mocks.Database{}
	cache := &mocks.Cache{}
	logger := zap.NewNop()

	svc := order_service.New(repo, cache, logger)
	ctx := context.Background()
	orderData := createValidOrder()

	repo.On("Create", ctx, mock.AnythingOfType("*models.Order")).Return(nil)
	cache.On("Set", ctx, "test-123", mock.AnythingOfType("*models.Order")).Return(nil)

	err := svc.ProcessOrder(ctx, orderData)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestOrderService_ProcessOrder_Duplicate(t *testing.T) {
	repo := &mocks.Database{}
	cache := &mocks.Cache{}
	logger := zap.NewNop()

	svc := order_service.New(repo, cache, logger)
	ctx := context.Background()
	orderData := createValidOrder()

	repo.On("Create", ctx, mock.AnythingOfType("*models.Order")).Return(errors.New("заказ уже существует в базе данных"))

	err := svc.ProcessOrder(ctx, orderData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "заказ уже существует в базе данных")
	repo.AssertExpectations(t)
	cache.AssertNotCalled(t, "Set")
}

func TestOrderService_ProcessOrder_Error_Cache(t *testing.T) {
	repo := &mocks.Database{}
	cache := &mocks.Cache{}
	logger := zap.NewNop()

	svc := order_service.New(repo, cache, logger)
	ctx := context.Background()
	orderData := createValidOrder()

	repo.On("Create", ctx, mock.AnythingOfType("*models.Order")).Return(errors.New("ошибка"))

	err := svc.ProcessOrder(ctx, orderData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repo.Create")
	repo.AssertExpectations(t)
	cache.AssertNotCalled(t, "Set")
}

// GetOrder Tests
func TestOrderSerivce_GetOrder_OK_FromDB(t *testing.T) {
	repo := &mocks.Database{}
	cache := &mocks.Cache{}
	logger := zap.NewNop()

	svc := order_service.New(repo, cache, logger)
	ctx := context.Background()
	expectedOrder := createValidOrder()

	cache.On("Get", ctx, "test-123").Return((*models.Order)(nil), errors.New("заказ не найден"))
	repo.On("Read", ctx, "test-123").Return(expectedOrder, nil)
	cache.On("Set", ctx, "test-123", expectedOrder).Return(nil)

	order, err := svc.GetOrder(ctx, "test-123")

	assert.NoError(t, err)
	assert.Equal(t, expectedOrder, order)
	cache.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestOrderSerivce_GetOrder_OK_FromCache(t *testing.T) {
	repo := &mocks.Database{}
	cache := &mocks.Cache{}
	logger := zap.NewNop()

	svc := order_service.New(repo, cache, logger)
	ctx := context.Background()
	expectedOrder := createValidOrder()

	cache.On("Get", ctx, "test-123").Return(expectedOrder, nil)

	order, err := svc.GetOrder(ctx, "test-123")

	assert.NoError(t, err)
	assert.Equal(t, expectedOrder, order)
	cache.AssertExpectations(t)
	repo.AssertNotCalled(t, "Read")
}

func TestOrderSerivce_GetOrder_NotFound(t *testing.T) {
	repo := &mocks.Database{}
	cache := &mocks.Cache{}
	logger := zap.NewNop()

	svc := order_service.New(repo, cache, logger)
	ctx := context.Background()

	cache.On("Get", ctx, "test-123").Return((*models.Order)(nil), errors.New("заказ не найден"))
	repo.On("Read", ctx, "test-123").Return((*models.Order)(nil), errors.New("заказ не найден"))

	order, err := svc.GetOrder(ctx, "test-123")

	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "заказ не найден")
	cache.AssertExpectations(t)
	repo.AssertExpectations(t)
}

// GetAllOrders Tests
func TestOrderSerivce_GetAllOrders_OK(t *testing.T) {
	repo := &mocks.Database{}
	cache := &mocks.Cache{}
	logger := zap.NewNop()

	svc := order_service.New(repo, cache, logger)
	ctx := context.Background()

	expectedOrders := []*models.Order{
		createValidOrder(),
		createValidOrder(),
	}
	expectedOrders[1].OrderUID = "test-456"

	repo.On("ReadAll", ctx).Return(expectedOrders, nil)
	cache.On("Restore", ctx, expectedOrders).Return(nil)

	orders, err := svc.GetAllOrders(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedOrders, orders)
	assert.Len(t, orders, 2)
	cache.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestOrderSerivce_GetAllOrders_OK_Empty(t *testing.T) {
	repo := &mocks.Database{}
	cache := &mocks.Cache{}
	logger := zap.NewNop()

	svc := order_service.New(repo, cache, logger)
	ctx := context.Background()

	repo.On("ReadAll", ctx).Return([]*models.Order{}, nil)
	cache.On("Restore", ctx, []*models.Order{}).Return(nil)

	orders, err := svc.GetAllOrders(ctx)

	assert.NoError(t, err)
	assert.Empty(t, orders)
	cache.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestOrderSerivce_GetAllOrders_Error_DB(t *testing.T) {
	repo := &mocks.Database{}
	cache := &mocks.Cache{}
	logger := zap.NewNop()

	svc := order_service.New(repo, cache, logger)
	ctx := context.Background()

	repo.On("ReadAll", ctx).Return(([]*models.Order)(nil), errors.New("ошибка БД"))

	orders, err := svc.GetAllOrders(ctx)

	assert.Error(t, err)
	assert.Nil(t, orders)
	assert.Contains(t, err.Error(), "repo.ReadAll")
	repo.AssertExpectations(t)
	cache.AssertNotCalled(t, "Restore")
}

func TestOrderSerivce_GetAllOrders_Error_Cache(t *testing.T) {
	repo := &mocks.Database{}
	cache := &mocks.Cache{}
	logger := zap.NewNop()

	svc := order_service.New(repo, cache, logger)
	ctx := context.Background()

	expectedOrders := []*models.Order{
		createValidOrder(),
		createValidOrder(),
		createValidOrder(),
	}
	expectedOrders[1].OrderUID = "test-456"
	expectedOrders[2].OrderUID = "test-789"

	repo.On("ReadAll", ctx).Return(expectedOrders, nil)
	cache.On("Restore", ctx, expectedOrders).Return(errors.New("ошибка восстановления кэша"))

	orders, err := svc.GetAllOrders(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedOrders, orders)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}
