package gocmdevt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"
)

type EventHandlerFunc func(ctx context.Context, cmd Event) (any, error)

// Event represents a domain event with common metadata
type Event interface {
	// EventID returns a unique identifier for this event instance
	EventID() string

	// EventType returns the type/name of the event
	EventType() string

	// EventTime returns when the event occurred
	EventTime() time.Time

	// AggregateID returns the ID of the aggregate that generated this event
	AggregateID() string

	// EventVersion returns the version of this event schema (for event evolution)
	EventVersion() int
}

// BaseEvent provides a default implementation of common event metadata
type BaseEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Time      time.Time `json:"time"`
	Aggregate string    `json:"aggregate_id"`
	Version   int       `json:"version"`
}

func NewBaseEvent(eventType, aggregateID string, version int) BaseEvent {
	return BaseEvent{
		ID:        generateEventID(),
		Type:      eventType,
		Time:      time.Now().UTC(),
		Aggregate: aggregateID,
		Version:   version,
	}
}

// generateEventID creates a unique identifier for events
func generateEventID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return time.Now().Format("20060102150405") + "-" + time.Now().Format("000000")
	}
	return hex.EncodeToString(bytes)
}

func (e BaseEvent) EventID() string {
	return e.ID
}

func (e BaseEvent) EventType() string {
	return e.Type
}

func (e BaseEvent) EventTime() time.Time {
	return e.Time
}

func (e BaseEvent) AggregateID() string {
	return e.Aggregate
}

func (e BaseEvent) EventVersion() int {
	return e.Version
}

// EventLogWriter is an interface for writing events to a log or database.
type EventLogWriter interface {
	Write(event Event) error
}

// Dispatcher is an interface for dispatching events to handlers
type Dispatcher interface {
	Dispatch(ctx context.Context, event Event)
}

// EventEmitter handles event emission with logging and dispatching
type EventEmitter struct {
	LogWriter  EventLogWriter
	Dispatcher Dispatcher
	// Queue       *QueuePublisher
}

func NewEventEmitter(logWriter EventLogWriter, dispatcher Dispatcher) *EventEmitter {
	return &EventEmitter{
		LogWriter:  logWriter,
		Dispatcher: dispatcher,
	}
}

func (e *EventEmitter) Emit(ctx context.Context, event Event) {
	// Log to DB
	if err := e.LogWriter.Write(event); err != nil {
		log.Printf("audit log failed: %v", err)
	}

	// In-process dispatch
	e.Dispatcher.Dispatch(ctx, event)

	// Optional async queue
	// e.Queue.Publish(event)
}
