package simpleapp

import (
	gocmdevt "github.com/leviplj/go-cmd-evt"
)

type OrderCreatedEvent struct {
	gocmdevt.BaseEvent
	CustomerID  string  `json:"customer_id"`
	ProductID   string  `json:"product_id"`
	Quantity    int     `json:"quantity"`
	TotalAmount float64 `json:"total_amount"`
}

func (e *OrderCreatedEvent) Payload() map[string]interface{} {
	return map[string]interface{}{
		"id":           e.ID,
		"type":         e.Type,
		"time":         e.Time,
		"customer_id":  e.CustomerID,
		"product_id":   e.ProductID,
		"quantity":     e.Quantity,
		"total_amount": e.TotalAmount,
	}
}

func NewOrderCreatedEvent(orderID, customerID, productID string, quantity int, totalAmount float64) *OrderCreatedEvent {
	return &OrderCreatedEvent{
		BaseEvent:   gocmdevt.NewBaseEvent("OrderCreated", orderID, 1),
		CustomerID:  customerID,
		ProductID:   productID,
		Quantity:    quantity,
		TotalAmount: totalAmount,
	}
}

type OrderShippedEvent struct {
	gocmdevt.BaseEvent
	OrderID         string `json:"order_id"`
	ShippingAddress string `json:"shipping_address"`
}

func (e *OrderShippedEvent) Payload() map[string]interface{} {
	return map[string]interface{}{
		"id":               e.ID,
		"type":             e.Type,
		"time":             e.Time,
		"order_id":         e.OrderID,
		"shipping_address": e.ShippingAddress,
	}
}

func NewOrderShippedEvent(orderID, shippingAddress string) *OrderShippedEvent {
	return &OrderShippedEvent{
		BaseEvent:       gocmdevt.NewBaseEvent("OrderShipped", orderID, 1),
		OrderID:         orderID,
		ShippingAddress: shippingAddress,
	}
}

type PaymentProcessedEvent struct {
	gocmdevt.BaseEvent
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	TransactionID string  `json:"transaction_id"`
}

func (e *PaymentProcessedEvent) Payload() map[string]interface{} {
	return map[string]interface{}{
		"id":             e.ID,
		"type":           e.Type,
		"time":           e.Time,
		"order_id":       e.OrderID,
		"amount":         e.Amount,
		"transaction_id": e.TransactionID,
	}
}

func NewPaymentProcessedEvent(orderID string, amount float64, transactionID string) *PaymentProcessedEvent {
	return &PaymentProcessedEvent{
		BaseEvent:     gocmdevt.NewBaseEvent("PaymentProcessed", orderID, 1),
		OrderID:       orderID,
		Amount:        amount,
		TransactionID: transactionID,
	}
}
