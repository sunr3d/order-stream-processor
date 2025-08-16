package http_handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/sunr3d/order-stream-processor/internal/handlers/validators"
	"github.com/sunr3d/order-stream-processor/internal/httpx"
)

func (h *httpHandler) createOrder(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("op", "handlers.createOrder"))

	logger.Info("получен запрос на создание заказа")

	var req createOrderReq

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		logger.Error("некорректный JSON", zap.Error(err))
		_ = httpx.HttpError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	if err := validators.ValidateOrder(&req); err != nil {
		logger.Error("ошибка валидации запроса", zap.Error(err))
		_ = httpx.HttpError(w, http.StatusBadRequest, err.Error())
		return
	}

	logger = logger.With(zap.String("order_uid", req.OrderUID))

	if err := h.svc.ProcessOrder(r.Context(), &req); err != nil {
		logger.Error("ошибка при обработке заказа", zap.Error(err))

		if strings.Contains(err.Error(), "уже существует") {
			logger.Info("заказ уже существует в БД")
			_ = httpx.HttpError(w, http.StatusConflict, "Заказ уже существует")
		} else {
			_ = httpx.HttpError(w, http.StatusInternalServerError, "Внутреняя ошибка сервера")
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
			_ = httpx.HttpError(w, http.StatusInternalServerError, "Внутреняя ошибка сервера")
		case errors.Is(err, httpx.ErrWriteBody):
			logger.Warn("клиент закрыл соединение, ответ не отправлен", zap.Error(err))
		}
		return
	}

	logger.Info("заказ успешно создан")
}

func (h *httpHandler) getOrder(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(zap.String("op", "handlers.getOrder"))

	logger.Info("получен запрос на получение заказа")

	orderUID := r.PathValue("order_uid")
	if orderUID == "" {
		logger.Error("пустой order_uid")
		_ = httpx.HttpError(w, http.StatusBadRequest, "order_uid не может быть пустым")
		return
	}
	logger = logger.With(zap.String("order_uid", orderUID))

	order, err := h.svc.GetOrder(r.Context(), orderUID)
	if err != nil {
		logger.Error("ошибка при получении заказа", zap.Error(err))

		if strings.Contains(err.Error(), "заказ не найден") {
			_ = httpx.HttpError(w, http.StatusNotFound, "Заказ не найден")
		} else {
			_ = httpx.HttpError(w, http.StatusInternalServerError, "Внутреняя ошибка сервера")
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
			_ = httpx.HttpError(w, http.StatusInternalServerError, "Внутреняя ошибка сервера")
		case errors.Is(err, httpx.ErrWriteBody):
			logger.Warn("клиент закрыл соединение, ответ не отправлен", zap.Error(err))
		}
		return
	}

	logger.Info("заказ успешно получен")
}
