package main

import (
	"context"
	"log"
	"reflect"

	gocmdevt "github.com/leviplj/go-cmd-evt"
)

type ConsoleEventLogger struct{}

func (l *ConsoleEventLogger) Write(event gocmdevt.Event) error {
	log.Printf("Event logged: %s at %s. ID: %s", event.EventType(), event.EventTime(), event.EventID())
	return nil
}

type InMemoryDispatcher struct {
	handlers map[reflect.Type][]gocmdevt.EventHandlerFunc
}

func NewInMemoryDispatcher() *InMemoryDispatcher {
	return &InMemoryDispatcher{
		handlers: make(map[reflect.Type][]gocmdevt.EventHandlerFunc),
	}
}

func (d *InMemoryDispatcher) Subscribe(event gocmdevt.Event, handler gocmdevt.EventHandlerFunc) {
	eventType := reflect.TypeOf(event)
	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

func (d *InMemoryDispatcher) Dispatch(event gocmdevt.Event) {
	ctx := context.Background()
	eventType := reflect.TypeOf(event)
	if handlers, exists := d.handlers[eventType]; exists {
		for _, handler := range handlers {
			handler(ctx, event)
		}
	}
}

func (d *InMemoryDispatcher) DispatchCtx(ctx context.Context, event gocmdevt.Event) {
	eventType := reflect.TypeOf(event)
	if handlers, exists := d.handlers[eventType]; exists {
		for _, handler := range handlers {
			handler(ctx, event)
		}
	}
}

type StudentModule struct {
	eventEmitter *gocmdevt.EventEmitter
}

func NewStudentModule(eventEmitter *gocmdevt.EventEmitter) *StudentModule {
	if eventEmitter == nil {
		panic("eventEmitter cannot be nil")
	}
	return &StudentModule{
		eventEmitter: eventEmitter,
	}
}

func (m *StudentModule) Handlers() map[reflect.Type]gocmdevt.HandlerFunc {
	return map[reflect.Type]gocmdevt.HandlerFunc{
		reflect.TypeOf(&CreateStudentCommand{}): m.createStudent,
	}
}

func (m *StudentModule) createStudent(ctx context.Context, cmd gocmdevt.Command) (any, error) {
	createCmd := cmd.(*CreateStudentCommand)

	// Create student logic here...

	// Emit event
	event := gocmdevt.NewBaseEvent("student_created", createCmd.ID, 1)
	m.eventEmitter.Emit(ctx, event)

	return nil, nil
}

type CreateStudentCommand struct {
	ID   string
	Name string
	Age  int
}

func main() {
	logger := &ConsoleEventLogger{}
	dispatcher := NewInMemoryDispatcher()

	// Create an event emitter
	eventEmitter := gocmdevt.NewEventEmitter(logger, dispatcher)

	// Create a student module with the event emitter
	studentModule := NewStudentModule(eventEmitter)

	// Create an application with the event emitter
	app := gocmdevt.NewApp(studentModule)

	// Example command to create a student
	createCmd := &CreateStudentCommand{
		ID:   "student-123",
		Name: "John Doe",
		Age:  20,
	}

	// Handle the command
	_, err := app.Handle(context.Background(), createCmd)
	if err != nil {
		log.Fatalf("Failed to handle command: %v", err)
	}
}
