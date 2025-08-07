package simpleapp

import (
	"context"
	"fmt"
	"reflect"

	gocmdevt "github.com/leviplj/go-cmd-evt"
)

type OrderModule struct {
	eventEmitter *gocmdevt.EventEmitter
}

func NewOrderModule(eventEmitter *gocmdevt.EventEmitter) *OrderModule {
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

	// Logic to create an order
	fmt.Printf("Creating order for customer %s with ID %s\n", createCmd.CustomerID, createCmd.OrderID)

	// Emit an event after creating the order
	event := NewOrderCreatedEvent(
		createCmd.OrderID,
		createCmd.CustomerID,
		createCmd.ProductID,
		createCmd.Quantity,
		createCmd.TotalAmount,
	)
	m.eventEmitter.Emit(event)

	return createCmd, nil
}

func (m *OrderModule) processPayment(ctx context.Context, cmd gocmdevt.Command) (any, error) {
	processCmd := cmd.(*ProcessPaymentCommand)

	// Logic to process payment
	fmt.Printf("Processing payment for order %s with amount %.2f\n", processCmd.OrderID, processCmd.Amount)

	// Emit an event after processing the payment
	event := NewPaymentProcessedEvent(
		processCmd.OrderID,
		processCmd.Amount,
		processCmd.TransactionID,
	)
	m.eventEmitter.Emit(event)

	return nil, nil
}

func (m *OrderModule) shipOrder(ctx context.Context, cmd gocmdevt.Command) (any, error) {
	shipCmd := cmd.(*ShipOrderCommand)

	// Logic to ship the order
	fmt.Printf("Shipping order %s to address %s\n", shipCmd.OrderID, shipCmd.ShippingAddress)

	// Emit an event after shipping the order
	event := NewOrderShippedEvent(
		shipCmd.OrderID,
		shipCmd.ShippingAddress,
	)
	m.eventEmitter.Emit(event)

	return nil, nil
}
