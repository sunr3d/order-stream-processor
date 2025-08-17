package validators

import (
	"fmt"
	"strings"

	"github.com/sunr3d/order-stream-processor/models"
)

func ValidateOrder(order *models.Order) error {
	// Основные поля
	if strings.TrimSpace(order.OrderUID) == "" {
		return fmt.Errorf("order_uid не может быть пустым")
	}
	if strings.TrimSpace(order.CustomerID) == "" {
		return fmt.Errorf("customer_id не может быть пустым")
	}
	if len(order.Items) == 0 {
		return fmt.Errorf("items не может быть пустым")
	}
	if strings.TrimSpace(order.TrackNumber) == "" {
		return fmt.Errorf("track_number не может быть пустым")
	}
	if order.DateCreated.IsZero() {
		return fmt.Errorf("date_created не может быть пустым")
	}

	// Поля доставки
	if strings.TrimSpace(order.Delivery.Name) == "" {
		return fmt.Errorf("delivery.name не может быть пустым")
	}
	if strings.TrimSpace(order.Delivery.Phone) == "" {
		return fmt.Errorf("delivery.phone не может быть пустым")
	}
	if strings.TrimSpace(order.Delivery.Email) == "" {
		return fmt.Errorf("delivery.email не может быть пустым")
	}
	if strings.TrimSpace(order.Delivery.City) == "" {
		return fmt.Errorf("delivery.city не может быть пустым")
	}
	if strings.TrimSpace(order.Delivery.Address) == "" {
		return fmt.Errorf("delivery.address не может быть пустым")
	}

	// Поля платежа
	if strings.TrimSpace(order.Payment.Transaction) == "" {
		return fmt.Errorf("payment.transaction не может быть пустым")
	}
	if strings.TrimSpace(order.Payment.Provider) == "" {
		return fmt.Errorf("payment.provider не может быть пустым")
	}
	if order.Payment.GoodsTotal <= 0 {
		return fmt.Errorf("payment.goods_total не может быть меньше или равно 0")
	}
	if order.Payment.DeliveryCost < 0 {
		return fmt.Errorf("payment.delivery_cost не может быть меньше 0")
	}
	if order.Payment.CustomFee < 0 {
		return fmt.Errorf("payment.custom_fee не может быть меньше 0")
	}
	if order.Payment.Amount <= 0 {
		return fmt.Errorf("payment.amount не может быть меньше или равно 0")
	}
	if order.Payment.PaymentDT <= 0 {
		return fmt.Errorf("payment.payment_dt не может быть меньше или равно 0")
	}

	return nil
}
