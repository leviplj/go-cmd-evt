package main

import (
	"context"
	"fmt"

	. "simple-app"

	gocmdevt "github.com/leviplj/go-cmd-evt"
)

func main() {
	fmt.Println("Starting Simple App...")

	// Initialize event emitter, dispatcher, and modules
	dispatcher := gocmdevt.NewInMemoryDispatcher()
	eventEmitter := gocmdevt.NewEventEmitter(
		gocmdevt.NewConsoleEventLogger(),
		dispatcher,
	)
	app := gocmdevt.NewApp()

	// Order module
	{
		orderModule := NewOrderModule(eventEmitter)
		app.RegisterModule(orderModule)

		// Register event handlers
		dispatcher.Subscribe(
			&OrderCreatedEvent{},
			func(ctx context.Context, cmd gocmdevt.Event) (any, error) {
				orderEvent := cmd.(*OrderCreatedEvent)
				fmt.Printf("[HANDLER] Order created: %v\n", orderEvent.Payload())

				processPaymentCmd := &ProcessPaymentCommand{
					OrderID:       orderEvent.ID,
					Amount:        orderEvent.TotalAmount,
					TransactionID: "txn-12345",
				}
				return app.Handle(ctx, processPaymentCmd)

				// shipOrderCmd := &ShipOrderCommand{
				// 	OrderID:         orderEvent.ID,
				// 	ShippingAddress: "123 Main St, Anytown, USA",
				// }
				// return app.Handle(ctx, shipOrderCmd)
			},
		)

		dispatcher.Subscribe(
			&PaymentProcessedEvent{},
			func(ctx context.Context, cmd gocmdevt.Event) (any, error) {
				paymentEvent := cmd.(*PaymentProcessedEvent)
				fmt.Printf("[HANDLER] Payment processed: %s for order %s\n", paymentEvent.TransactionID, paymentEvent.OrderID)

				shipOrderCmd := &ShipOrderCommand{
					OrderID:         paymentEvent.OrderID,
					ShippingAddress: "123 Main St, Anytown, USA",
				}
				return app.Handle(ctx, shipOrderCmd)
			},
		)

		dispatcher.Subscribe(
			&OrderShippedEvent{},
			func(ctx context.Context, cmd gocmdevt.Event) (any, error) {
				shipEvent := cmd.(*OrderShippedEvent)
				fmt.Printf("[HANDLER] Order shipped: %s with address %s\n", shipEvent.OrderID, shipEvent.ShippingAddress)
				return nil, nil
			},
		)

	}

	newOrderCmd := &CreateOrderCommand{
		OrderID:     "order-123",
		CustomerID:  "customer-456",
		ProductID:   "product-789",
		Quantity:    2,
		TotalAmount: 199.98,
	}

	res, err := app.Handle(context.TODO(), newOrderCmd)
	if err != nil {
		fmt.Printf("Error handling command: %v\n", err)
		return
	}
	_ = res // Use the result as needed

	// fmt.Printf("Order created successfully: %v\n", res)
}
