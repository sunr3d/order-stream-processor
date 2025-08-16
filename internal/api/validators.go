package api

import (
	"fmt"
	"strings"

	"github.com/sunr3d/order-stream-processor/models"
)

// Валидация заказа (только основные поля, можно расширить по мере необходимости)
func validateCreateOrderReq(req models.Order) error {
	// Основные поля
	if strings.TrimSpace(req.OrderUID) == "" {
		return fmt.Errorf("order_uid не может быть пустым")
	}
	if strings.TrimSpace(req.CustomerID) == "" {
		return fmt.Errorf("customer_id не может быть пустым")
	}
	if len(req.Items) == 0 {
		return fmt.Errorf("items не может быть пустым")
	}
	if strings.TrimSpace(req.TrackNumber) == "" {
		return fmt.Errorf("track_number не может быть пустым")
	}

	// Поля доставки
	if strings.TrimSpace(req.Delivery.Name) == "" {
		return fmt.Errorf("delivery.name не может быть пустым")
	}

	// Поля платежа
	if strings.TrimSpace(req.Payment.Transaction) == "" {
		return fmt.Errorf("payment.transaction не может быть пустым")
	}
	if strings.TrimSpace(req.Payment.Provider) == "" {
		return fmt.Errorf("payment.provider не может быть пустым")
	}
	if req.Payment.Amount <= 0 {
		return fmt.Errorf("payment.amount не может быть меньше или равно 0")
	}
	if req.Payment.PaymentDT <= 0 {
		return fmt.Errorf("payment.payment_dt не может быть меньше или равно 0")
	}

	return nil
}
