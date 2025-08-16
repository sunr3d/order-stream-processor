package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/sunr3d/order-stream-processor/internal/api"
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

// createOrder Handler Tests
func TestHandler_CreateOrder_OK(t *testing.T) {
	svc := &mocks.OrderService{}
	logger := zap.NewNop()
	controller := api.New(svc, logger)

	orderData := createValidOrder()
	jsonData, err := json.Marshal(orderData)
	assert.NoError(t, err)

	svc.On("ProcessOrder", mock.Anything, mock.AnythingOfType("*models.Order")).Return(nil)

	mux := http.NewServeMux()
	controller.RegisterOrderHandlers(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Post(server.URL+"/order", "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var respJSON map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respJSON)
	assert.NoError(t, err)

	assert.Equal(t, "test-123", respJSON["order_uid"])
	assert.Equal(t, "Заказ успешно создан", respJSON["message"])

	svc.AssertExpectations(t)
}

func TestHandler_CreateOrder_Error_InvalidJSON(t *testing.T) {
	svc := &mocks.OrderService{}
	logger := zap.NewNop()
	controller := api.New(svc, logger)

	mux := http.NewServeMux()
	controller.RegisterOrderHandlers(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Post(server.URL+"/order", "application/json", bytes.NewBufferString(`{"invalid": json`))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respJSON map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respJSON)
	assert.NoError(t, err)
	assert.Equal(t, "Некорректный JSON", respJSON["error"])

	svc.AssertNotCalled(t, "ProcessOrder")
}

func TestHandler_CreateOrder_Error_Validation(t *testing.T) {
	svc := &mocks.OrderService{}
	logger := zap.NewNop()
	controller := api.New(svc, logger)

	orderData := createValidOrder()
	orderData.OrderUID = ""
	jsonData, _ := json.Marshal(orderData)

	mux := http.NewServeMux()
	controller.RegisterOrderHandlers(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Post(server.URL+"/order", "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var respJSON map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respJSON)
	assert.NoError(t, err)
	assert.Equal(t, "order_uid не может быть пустым", respJSON["error"])

	svc.AssertNotCalled(t, "ProcessOrder")
}

func TestHandler_CreateOrder_Error_Duplicate(t *testing.T) {
	svc := &mocks.OrderService{}
	logger := zap.NewNop()
	controller := api.New(svc, logger)

	orderData := createValidOrder()
	jsonData, _ := json.Marshal(orderData)

	svc.On("ProcessOrder", mock.Anything, mock.AnythingOfType("*models.Order")).Return(errors.New("заказ уже существует"))

	mux := http.NewServeMux()
	controller.RegisterOrderHandlers(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Post(server.URL+"/order", "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var respJSON map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respJSON)
	assert.NoError(t, err)
	assert.Equal(t, "Заказ уже существует", respJSON["error"])

	svc.AssertExpectations(t)
}

// getOrder Handler Tests
func TestHandler_GetOrder_OK(t *testing.T) {
	svc := &mocks.OrderService{}
	logger := zap.NewNop()
	controller := api.New(svc, logger)

	expectedOrder := createValidOrder()

	svc.On("GetOrder", mock.Anything, "test-123").Return(expectedOrder, nil)

	mux := http.NewServeMux()
	controller.RegisterOrderHandlers(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/order/test-123")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respJSON map[string]any
	err = json.NewDecoder(resp.Body).Decode(&respJSON)
	assert.NoError(t, err)

	order, exists := respJSON["order"]
	assert.True(t, exists)

	orderMap := order.(map[string]any)
	assert.Equal(t, expectedOrder.OrderUID, orderMap["order_uid"])

	svc.AssertExpectations(t)
}

func TestHandler_GetOrder_Error_NotFound(t *testing.T) {
	svc := &mocks.OrderService{}
	logger := zap.NewNop()
	controller := api.New(svc, logger)

	svc.On("GetOrder", mock.Anything, "test-123").Return((*models.Order)(nil), errors.New("заказ не найден"))

	mux := http.NewServeMux()
	controller.RegisterOrderHandlers(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/order/test-123")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var respJSON map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respJSON)
	assert.NoError(t, err)
	assert.Equal(t, "Заказ не найден", respJSON["error"])

	svc.AssertExpectations(t)
}
