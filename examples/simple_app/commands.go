package simpleapp

type CreateOrderCommand struct {
	OrderID     string
	CustomerID  string
	ProductID   string
	Quantity    int
	TotalAmount float64
}

type ProcessPaymentCommand struct {
	OrderID       string
	Amount        float64
	TransactionID string
}

type ShipOrderCommand struct {
	OrderID         string
	ShippingAddress string
}
