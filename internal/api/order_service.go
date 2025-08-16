package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/sunr3d/order-stream-processor/internal/httpx"
	"github.com/sunr3d/order-stream-processor/internal/interfaces/services"
)

type handler struct {
	svc    services.OrderService
	logger *zap.Logger
}

func New(svc services.OrderService, logger *zap.Logger) *handler {
	return &handler{svc: svc, logger: logger}
}

func (h *handler) RegisterOrderHandlers(mux *http.ServeMux) {
	mux.HandleFunc("POST /order", h.createOrder)
	mux.HandleFunc("GET /order/{order_uid}", h.getOrder)
}

func (h *handler) createOrder(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("op", "handlers.createOrder"))

	logger.Info("получен запрос на создание заказа")

	var req createOrderReq

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		logger.Error("некорректный JSON", zap.Error(err))
		httpx.HttpError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	if err := validateCreateOrderReq(req); err != nil {
		logger.Error("ошибка валидации запроса", zap.Error(err))
		httpx.HttpError(w, http.StatusBadRequest, err.Error())
		return
	}

	logger = logger.With(zap.String("order_uid", req.OrderUID))

	if err := h.svc.ProcessOrder(r.Context(), &req); err != nil {
		logger.Error("ошибка при обработке заказа", zap.Error(err))

		if strings.Contains(err.Error(), "уже существует") {
			logger.Info("заказ уже существует в БД")
			httpx.HttpError(w, http.StatusConflict, "Заказ уже существует")
		} else {
			httpx.HttpError(w, http.StatusInternalServerError, "Внутреняя ошибка сервера")
		}
		return
	}

	resp := createOrderResp{
		OrderUID: req.OrderUID,
		Message:  "Заказ успешно создан",
	}

	if err := httpx.WriteJSON(w, http.StatusCreated, resp); err != nil {
		switch {
		case errors.Is(err, httpx.ErrJSONMarshal):
			logger.Error("ошибка при отправке ответа", zap.Error(err))
			httpx.HttpError(w, http.StatusInternalServerError, "Внутреняя ошибка сервера")
		case errors.Is(err, httpx.ErrWriteBody):
			logger.Warn("клиент закрыл соединение, ответ не отправлен", zap.Error(err))
		}
		return
	}

	logger.Info("заказ успешно создан")
}

func (h *handler) getOrder(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("op", "handlers.getOrder"))

	logger.Info("получен запрос на получение заказа")

	orderUID := r.PathValue("order_uid")
	if orderUID == "" {
		logger.Error("пустой order_uid")
		httpx.HttpError(w, http.StatusBadRequest, "order_uid не может быть пустым")
		return
	}
	logger = logger.With(zap.String("order_uid", orderUID))

	order, err := h.svc.GetOrder(r.Context(), orderUID)
	if err != nil {
		logger.Error("ошибка при получении заказа", zap.Error(err))

		if strings.Contains(err.Error(), "заказ не найден") {
			httpx.HttpError(w, http.StatusNotFound, "Заказ не найден")
		} else {
			httpx.HttpError(w, http.StatusInternalServerError, "Внутреняя ошибка сервера")
		}
		return
	}

	resp := getOrderResp{
		Order: order,
	}

	if err := httpx.WriteJSON(w, http.StatusOK, resp); err != nil {
		switch {
		case errors.Is(err, httpx.ErrJSONMarshal):
			logger.Error("ошибка при отправке ответа", zap.Error(err))
			httpx.HttpError(w, http.StatusInternalServerError, "Внутреняя ошибка сервера")
		case errors.Is(err, httpx.ErrWriteBody):
			logger.Warn("клиент закрыл соединение, ответ не отправлен", zap.Error(err))
		}
		return
	}

	logger.Info("заказ успешно получен")
}
