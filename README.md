# go-cmd-evt

A lightweight Go library for implementing Command Query Responsibility Segregation (CQRS) and Event Sourcing patterns with modular architecture support.

## Overview

`go-cmd-evt` provides a simple yet powerful framework for building applications using:
- **Command Pattern**: Execute operations through command handlers
- **Event Sourcing**: Emit and handle domain events
- **Modular Architecture**: Organize code into reusable modules
- **Context-Aware Processing**: Full context propagation through command and event handling

## Features

- **Command Handling**: Type-safe command routing and execution
- **Event System**: Complete event lifecycle management with metadata
- **Module System**: Organize handlers into logical modules
- **Event Logging**: Pluggable event logging infrastructure
- **In-Memory Dispatching**: Fast, reflection-based event dispatching
- **Context Support**: Context propagation for cancellation and deadlines

## Installation

```bash
go get github.com/leviplj/go-cmd-evt
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    gocmdevt "github.com/leviplj/go-cmd-evt"
)

// Define a command
type CreateUserCommand struct {
    UserID string
    Name   string
}

// Define an event
type UserCreatedEvent struct {
    gocmdevt.BaseEvent
    UserID string
    Name   string
}

// Create a module
type UserModule struct {
    eventEmitter *gocmdevt.EventEmitter
}

func (m *UserModule) Handlers() map[reflect.Type]gocmdevt.HandlerFunc {
    return map[reflect.Type]gocmdevt.HandlerFunc{
        reflect.TypeOf(&CreateUserCommand{}): m.createUser,
    }
}

func (m *UserModule) createUser(ctx context.Context, cmd gocmdevt.Command) (any, error) {
    createCmd := cmd.(*CreateUserCommand)

    // Business logic here
    fmt.Printf("Creating user: %s\n", createCmd.Name)

    // Emit event
    event := &UserCreatedEvent{
        BaseEvent: gocmdevt.NewBaseEvent("UserCreated", createCmd.UserID, 1),
        UserID:    createCmd.UserID,
        Name:      createCmd.Name,
    }
    m.eventEmitter.Emit(event)

    return nil, nil
}

func main() {
    // Initialize the application
    dispatcher := gocmdevt.NewInMemoryDispatcher()
    eventEmitter := gocmdevt.NewEventEmitter(
        gocmdevt.NewConsoleEventLogger(),
        dispatcher,
    )
    app := gocmdevt.NewApp()

    // Register module
    userModule := &UserModule{eventEmitter: eventEmitter}
    app.RegisterModule(userModule)

    // Handle command
    cmd := &CreateUserCommand{UserID: "123", Name: "John Doe"}
    _, err := app.Handle(context.Background(), cmd)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Core Components

### Commands

Commands represent intent to perform an action. They are simple structs that implement the `Command` interface (which is just an empty marker interface).

```go
type Command interface{}

type YourCommand struct {
    Field1 string
    Field2 int
}
```

### Events

Events represent something that has happened in the system. They must implement the `Event` interface:

```go
type Event interface {
    EventID() string
    EventType() string
    EventTime() time.Time
    AggregateID() string
    EventVersion() int
    Payload() map[string]interface{}
}
```

The library provides `BaseEvent` as a convenient base implementation:

```go
type YourEvent struct {
    gocmdevt.BaseEvent
    CustomField string
}

func NewYourEvent(aggregateID, customField string) *YourEvent {
    return &YourEvent{
        BaseEvent:   gocmdevt.NewBaseEvent("YourEventType", aggregateID, 1),
        CustomField: customField,
    }
}
```

### Modules

Modules organize related command handlers into logical units:

```go
type Module interface {
    Handlers() map[reflect.Type]HandlerFunc
}
```

### Event Emitter

The `EventEmitter` coordinates event logging and dispatching:

```go
emitter := gocmdevt.NewEventEmitter(logger, dispatcher)

// Emit events
emitter.Emit(event)

// Or with context
emitter.EmitCtx(ctx, event)
```

### Event Dispatcher

The dispatcher routes events to registered handlers:

```go
dispatcher := gocmdevt.NewInMemoryDispatcher()

// Subscribe to events
dispatcher.Subscribe(&YourEvent{}, func(ctx context.Context, evt Event) (any, error) {
    // Handle event
    return nil, nil
})

// Dispatch events
dispatcher.Dispatch(event)
```

## Complete Example

See the `/examples/simple_app` directory for a complete order processing system demonstrating:

- Command handling for order creation, payment processing, and shipping
- Event emission and handling with automatic workflow progression
- Module organization for order management
- Chain of events triggering subsequent commands

### Running the Example

```bash
cd examples/simple_app
go run cmd/main.go
```

The example demonstrates a complete order workflow:
1. Create order → emits `OrderCreatedEvent`
2. Handle `OrderCreatedEvent` → triggers payment processing
3. Process payment → emits `PaymentProcessedEvent`
4. Handle `PaymentProcessedEvent` → triggers shipping
5. Ship order → emits `OrderShippedEvent`

## Architecture Patterns

### CQRS Pattern

The library naturally supports CQRS by separating:
- **Commands**: Write operations handled by command handlers
- **Events**: State changes propagated through the system
- **Queries**: Read operations (implement separately as needed)

### Event Sourcing

Store events as the source of truth for system state:
1. Commands trigger business logic
2. Business logic emits events
3. Events are persisted (via `EventLogWriter`)
4. State can be reconstructed from events

### Modular Monolith

Organize your application into modules:
- Each module handles a specific domain
- Modules are loosely coupled through events
- Easy to split into microservices later

## Customization

### Custom Event Logger

Implement the `EventLogWriter` interface:

```go
type DatabaseEventLogger struct {
    db *sql.DB
}

func (l *DatabaseEventLogger) Write(event Event) error {
    // Save event to database
    return nil
}
```

### Custom Dispatcher

Implement the `Dispatcher` interface for custom routing logic:

```go
type QueueDispatcher struct {
    queue MessageQueue
}

func (d *QueueDispatcher) Dispatch(event Event) {
    // Send to message queue
}

func (d *QueueDispatcher) DispatchCtx(ctx context.Context, event Event) {
    // Send to message queue with context
}
```

## Best Practices

1. **Keep Commands Simple**: Commands should only contain data, no logic
2. **Immutable Events**: Events represent history and should never be modified
3. **Single Responsibility**: Each handler should do one thing well
4. **Use Context**: Always propagate context for proper cancellation
5. **Event Versioning**: Use version numbers for event schema evolution
6. **Idempotent Handlers**: Design handlers to be safely retryable

## Testing

The library is designed for easy testing:

```go
func TestOrderCreation(t *testing.T) {
    // Create test dependencies
    dispatcher := gocmdevt.NewInMemoryDispatcher()
    logger := &MockEventLogger{}
    emitter := gocmdevt.NewEventEmitter(logger, dispatcher)

    // Create module
    module := NewOrderModule(emitter)

    // Test command handling
    cmd := &CreateOrderCommand{OrderID: "123"}
    result, err := module.createOrder(context.Background(), cmd)

    // Assert results
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

[MIT License](LICENSE)