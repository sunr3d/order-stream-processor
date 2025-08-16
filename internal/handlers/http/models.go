package http_handlers

import (
	"github.com/sunr3d/order-stream-processor/models"
)

type createOrderReq = models.Order

type createOrderResp struct {
	OrderUID string `json:"order_uid"`
	Message  string `json:"message"`
}

type getOrderResp struct {
	Order *models.Order `json:"order"`
}
