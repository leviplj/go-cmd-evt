package gocmdevt_test

import (
	"context"
	"fmt"
	"log"
	"reflect"

	gocmdevt "github.com/leviplj/go-cmd-evt"
)

// Example command types for a simple e-commerce system
type CreateOrderCommand struct {
	CustomerID  string
	ProductID   string
	Quantity    int
	TotalAmount float64
}

type ProcessPaymentCommand struct {
	OrderID string
	Amount  float64
}

type ShipOrderCommand struct {
	OrderID         string
	ShippingAddress string
}

// Example event types
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

type PaymentProcessedEvent struct {
	gocmdevt.BaseEvent
	Amount        float64 `json:"amount"`
	TransactionID string  `json:"transaction_id"`
}

func NewPaymentProcessedEvent(orderID string, amount float64, transactionID string) *PaymentProcessedEvent {
	return &PaymentProcessedEvent{
		BaseEvent:     gocmdevt.NewBaseEvent("PaymentProcessed", orderID, 1),
		Amount:        amount,
		TransactionID: transactionID,
	}
}

type OrderShippedEvent struct {
	gocmdevt.BaseEvent
	TrackingNumber  string `json:"tracking_number"`
	ShippingAddress string `json:"shipping_address"`
}

func NewOrderShippedEvent(orderID, trackingNumber, shippingAddress string) *OrderShippedEvent {
	return &OrderShippedEvent{
		BaseEvent:       gocmdevt.NewBaseEvent("OrderShipped", orderID, 1),
		TrackingNumber:  trackingNumber,
		ShippingAddress: shippingAddress,
	}
}

// Simple implementations of EventLogWriter and Dispatcher
type ConsoleEventLogger struct{}

func (c *ConsoleEventLogger) Write(event gocmdevt.Event) error {
	fmt.Printf("[EVENT LOG] %s for %s\n", event.EventType(), event.AggregateID())
	return nil
}

type InMemoryDispatcher struct {
	handlers map[reflect.Type][]func(event gocmdevt.Event)
}

func NewInMemoryDispatcher() *InMemoryDispatcher {
	return &InMemoryDispatcher{
		handlers: make(map[reflect.Type][]func(event gocmdevt.Event)),
	}
}

func (d *InMemoryDispatcher) Subscribe(eventType reflect.Type, handler func(event gocmdevt.Event)) {
	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

func (d *InMemoryDispatcher) Dispatch(ctx context.Context, event gocmdevt.Event) {
	eventType := reflect.TypeOf(event)
	if handlers, exists := d.handlers[eventType]; exists {
		for _, handler := range handlers {
			handler(event)
		}
	}
}

// Example module that handles order-related commands
type OrderModule struct {
	eventEmitter *gocmdevt.EventEmitter
}

func NewOrderModule(eventEmitter *gocmdevt.EventEmitter) *OrderModule {
	if eventEmitter == nil {
		panic("eventEmitter cannot be nil")
	}
	return &OrderModule{
		eventEmitter: eventEmitter,
	}
}

func (m *OrderModule) Handlers() map[reflect.Type]gocmdevt.HandlerFunc {
	return map[reflect.Type]gocmdevt.HandlerFunc{
		reflect.TypeOf(&CreateOrderCommand{}):    m.createOrder,
		reflect.TypeOf(&ProcessPaymentCommand{}): m.processPayment,
		reflect.TypeOf(&ShipOrderCommand{}):      m.shipOrder,
	}
}

func (m *OrderModule) createOrder(ctx context.Context, cmd gocmdevt.Command) (any, error) {
	createCmd := cmd.(*CreateOrderCommand)

	// Simulate order creation logic
	orderID := fmt.Sprintf("order-%s-%s", createCmd.CustomerID, createCmd.ProductID)

	result := map[string]interface{}{
		"order_id":     orderID,
		"customer_id":  createCmd.CustomerID,
		"product_id":   createCmd.ProductID,
		"quantity":     createCmd.Quantity,
		"total_amount": createCmd.TotalAmount,
		"status":       "created",
	}

	// Emit event after successful command handling
	event := NewOrderCreatedEvent(orderID, createCmd.CustomerID, createCmd.ProductID, createCmd.Quantity, createCmd.TotalAmount)
	m.eventEmitter.Emit(context.TODO(), event)

	return result, nil
}

func (m *OrderModule) processPayment(ctx context.Context, cmd gocmdevt.Command) (any, error) {
	paymentCmd := cmd.(*ProcessPaymentCommand)

	// Simulate payment processing
	if paymentCmd.Amount <= 0 {
		return nil, fmt.Errorf("invalid payment amount: %f", paymentCmd.Amount)
	}

	transactionID := fmt.Sprintf("txn-%s", paymentCmd.OrderID)
	result := map[string]interface{}{
		"order_id":       paymentCmd.OrderID,
		"amount_charged": paymentCmd.Amount,
		"payment_status": "completed",
		"transaction_id": transactionID,
	}

	// Emit event after successful payment processing
	event := NewPaymentProcessedEvent(paymentCmd.OrderID, paymentCmd.Amount, transactionID)
	m.eventEmitter.Emit(context.TODO(), event)

	return result, nil
}

func (m *OrderModule) shipOrder(ctx context.Context, cmd gocmdevt.Command) (any, error) {
	shipCmd := cmd.(*ShipOrderCommand)

	trackingNumber := fmt.Sprintf("TRK%s", shipCmd.OrderID)
	result := map[string]interface{}{
		"order_id":         shipCmd.OrderID,
		"shipping_address": shipCmd.ShippingAddress,
		"tracking_number":  trackingNumber,
		"status":           "shipped",
	}

	// Emit event after successful shipping
	event := NewOrderShippedEvent(shipCmd.OrderID, trackingNumber, shipCmd.ShippingAddress)
	m.eventEmitter.Emit(context.TODO(), event)

	return result, nil
}

// Example demonstrates how to use the go-cmd-evt library with events
func Example() {
	// Set up event handling
	logger := &ConsoleEventLogger{}
	dispatcher := NewInMemoryDispatcher()

	// Subscribe to events
	dispatcher.Subscribe(reflect.TypeOf(&OrderCreatedEvent{}), func(event gocmdevt.Event) {
		orderEvent := event.(*OrderCreatedEvent)
		fmt.Printf("[HANDLER] Order created notification sent to customer %s\n", orderEvent.CustomerID)
	})

	dispatcher.Subscribe(reflect.TypeOf(&PaymentProcessedEvent{}), func(event gocmdevt.Event) {
		paymentEvent := event.(*PaymentProcessedEvent)
		fmt.Printf("[HANDLER] Payment confirmation email sent for order %s\n", paymentEvent.AggregateID())
	})

	dispatcher.Subscribe(reflect.TypeOf(&OrderShippedEvent{}), func(event gocmdevt.Event) {
		shippedEvent := event.(*OrderShippedEvent)
		fmt.Printf("[HANDLER] Shipping notification sent with tracking %s\n", shippedEvent.TrackingNumber)
	})

	// Create event emitter
	eventEmitter := gocmdevt.NewEventEmitter(logger, dispatcher)

	// Create the application with events and modules
	orderModule := NewOrderModule(eventEmitter)
	app := gocmdevt.NewApp(orderModule)

	ctx := context.Background()

	// Example 1: Create an order
	createOrderCmd := &CreateOrderCommand{
		CustomerID:  "customer-123",
		ProductID:   "product-456",
		Quantity:    2,
		TotalAmount: 99.99,
	}

	result, err := app.Handle(ctx, createOrderCmd)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		return
	}

	orderResult := result.(map[string]interface{})
	fmt.Printf("Order created: %s\n", orderResult["order_id"])

	// Example 2: Process payment for the order
	processPaymentCmd := &ProcessPaymentCommand{
		OrderID: orderResult["order_id"].(string),
		Amount:  99.99,
	}

	paymentResult, err := app.Handle(ctx, processPaymentCmd)
	if err != nil {
		log.Printf("Error processing payment: %v", err)
		return
	}

	payment := paymentResult.(map[string]interface{})
	fmt.Printf("Payment processed: %s\n", payment["transaction_id"])

	// Example 3: Ship the order
	shipOrderCmd := &ShipOrderCommand{
		OrderID:         orderResult["order_id"].(string),
		ShippingAddress: "123 Main St, City, State 12345",
	}

	shipResult, err := app.Handle(ctx, shipOrderCmd)
	if err != nil {
		log.Printf("Error shipping order: %v", err)
		return
	}

	shipping := shipResult.(map[string]interface{})
	fmt.Printf("Order shipped: %s\n", shipping["tracking_number"])

	// Output:
	// [EVENT LOG] OrderCreated for order-customer-123-product-456
	// [HANDLER] Order created notification sent to customer customer-123
	// Order created: order-customer-123-product-456
	// [EVENT LOG] PaymentProcessed for order-customer-123-product-456
	// [HANDLER] Payment confirmation email sent for order order-customer-123-product-456
	// Payment processed: txn-order-customer-123-product-456
	// [EVENT LOG] OrderShipped for order-customer-123-product-456
	// [HANDLER] Shipping notification sent with tracking TRKorder-customer-123-product-456
	// Order shipped: TRKorder-customer-123-product-456
}

// ExampleOrderModule_nilEventEmitter demonstrates that the constructor panics when nil is passed
func ExampleOrderModule_nilEventEmitter() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic caught: %v\n", r)
		}
	}()

	// This should panic
	NewOrderModule(nil)

	// Output:
	// Panic caught: eventEmitter cannot be nil
}

// ExampleEvent_metadata demonstrates the improved Event interface
func ExampleEvent_metadata() {
	event := NewOrderCreatedEvent("order-123", "customer-456", "product-789", 2, 99.99)

	fmt.Printf("Event Type: %s\n", event.EventType())
	fmt.Printf("Aggregate ID: %s\n", event.AggregateID())
	fmt.Printf("Event Version: %d\n", event.EventVersion())
	fmt.Printf("Has Event ID: %t\n", len(event.EventID()) > 0)
	fmt.Printf("Has Timestamp: %t\n", !event.EventTime().IsZero())
	fmt.Printf("Customer ID: %s\n", event.CustomerID)

	// Output:
	// Event Type: OrderCreated
	// Aggregate ID: order-123
	// Event Version: 1
	// Has Event ID: true
	// Has Timestamp: true
	// Customer ID: customer-456
}
