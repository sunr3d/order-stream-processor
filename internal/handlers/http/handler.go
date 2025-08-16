package http_handlers

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/sunr3d/order-stream-processor/internal/interfaces/services"
)

// Структура HTTP обработчика
type httpHandler struct {
	svc    services.OrderService
	logger *zap.Logger
}

func New(svc services.OrderService, logger *zap.Logger) *httpHandler {
	return &httpHandler{svc: svc, logger: logger}
}

func (h *httpHandler) RegisterOrderHandlers(mux *http.ServeMux) {
	mux.HandleFunc("POST /order", h.createOrder)
	mux.HandleFunc("GET /order/{order_uid}", h.getOrder)
	mux.HandleFunc("GET /health", h.healthCheck)
}
