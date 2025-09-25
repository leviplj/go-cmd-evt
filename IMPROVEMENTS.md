# go-cmd-evt Improvement Recommendations

## Critical Issues

### 1. Thread Safety
**Problem:** The `InMemoryDispatcher` lacks synchronization for concurrent access to the handlers map, which could lead to race conditions and panics in production.

**Solution:**
```go
type InMemoryDispatcher struct {
    mu       sync.RWMutex
    handlers map[reflect.Type][]EventHandlerFunc
}

func (d *InMemoryDispatcher) Subscribe(event Event, handler EventHandlerFunc) {
    d.mu.Lock()
    defer d.mu.Unlock()
    eventType := reflect.TypeOf(event)
    d.handlers[eventType] = append(d.handlers[eventType], handler)
}

func (d *InMemoryDispatcher) DispatchCtx(ctx context.Context, event Event) {
    d.mu.RLock()
    defer d.mu.RUnlock()
    // ... dispatch logic
}
```

### 2. Error Handling
**Problem:** Event handler errors are ignored, making debugging difficult and potentially losing critical failure information.

**Solution:**
- Add error callback mechanism
- Implement retry logic with exponential backoff
- Create error aggregation for batch operations

```go
type EventEmitter struct {
    LogWriter    EventLogWriter
    Dispatcher   Dispatcher
    ErrorHandler func(event Event, err error)
}

func (e *EventEmitter) EmitCtx(ctx context.Context, event Event) error {
    if err := e.LogWriter.Write(event); err != nil {
        if e.ErrorHandler != nil {
            e.ErrorHandler(event, err)
        }
        return fmt.Errorf("failed to write event log: %w", err)
    }

    if err := e.Dispatcher.DispatchCtx(ctx, event); err != nil {
        if e.ErrorHandler != nil {
            e.ErrorHandler(event, err)
        }
        return fmt.Errorf("failed to dispatch event: %w", err)
    }

    return nil
}
```

## Performance Improvements

### 3. Event ID Generation
**Problem:** Using crypto/rand for every event ID is computationally expensive for high-throughput systems.

**Solution:**
- Use UUID v7 (time-ordered) for better performance and indexing
- Implement ID generator interface for flexibility
- Consider using xid or ulid libraries

```go
type EventIDGenerator interface {
    Generate() string
}

type ULIDGenerator struct{}

func (g *ULIDGenerator) Generate() string {
    return ulid.MustNew(ulid.Now(), entropy).String()
}
```

### 4. Event Batching
**Problem:** No support for batching events, leading to inefficient I/O operations.

**Solution:**
```go
type BatchEventEmitter struct {
    *EventEmitter
    batch      []Event
    batchSize  int
    flushTimer *time.Timer
    mu         sync.Mutex
}

func (e *BatchEventEmitter) EmitBatch(events []Event) error {
    e.mu.Lock()
    defer e.mu.Unlock()

    e.batch = append(e.batch, events...)
    if len(e.batch) >= e.batchSize {
        return e.flush()
    }
    return nil
}
```

## Feature Additions

### 5. Middleware Support
**Problem:** No way to add cross-cutting concerns like logging, metrics, or authentication.

**Solution:**
```go
type Middleware func(next HandlerFunc) HandlerFunc

func LoggingMiddleware() Middleware {
    return func(next HandlerFunc) HandlerFunc {
        return func(ctx context.Context, cmd Command) (any, error) {
            start := time.Now()
            result, err := next(ctx, cmd)
            log.Printf("Command %T took %v", cmd, time.Since(start))
            return result, err
        }
    }
}

type App struct {
    handlers    map[reflect.Type]HandlerFunc
    middlewares []Middleware
}

func (a *App) Use(mw Middleware) {
    a.middlewares = append(a.middlewares, mw)
}
```

### 6. Event Store Abstraction
**Problem:** No persistence layer abstraction for event sourcing.

**Solution:**
```go
type EventStore interface {
    Save(ctx context.Context, events ...Event) error
    Load(ctx context.Context, aggregateID string, fromVersion int) ([]Event, error)
    LoadSnapshot(ctx context.Context, aggregateID string) (*Snapshot, error)
    SaveSnapshot(ctx context.Context, snapshot Snapshot) error
}

type Snapshot struct {
    AggregateID string
    Version     int
    Data        []byte
    Timestamp   time.Time
}
```

### 7. Saga/Process Manager Support
**Problem:** No built-in support for long-running business processes.

**Solution:**
```go
type Saga interface {
    Handle(ctx context.Context, event Event) ([]Command, error)
    IsComplete() bool
    Compensate(ctx context.Context) error
}

type SagaManager struct {
    sagas map[string]Saga
    app   *App
}

func (sm *SagaManager) HandleEvent(ctx context.Context, event Event) error {
    saga, exists := sm.sagas[event.AggregateID()]
    if !exists {
        return nil
    }

    commands, err := saga.Handle(ctx, event)
    if err != nil {
        return saga.Compensate(ctx)
    }

    for _, cmd := range commands {
        if _, err := sm.app.Handle(ctx, cmd); err != nil {
            return saga.Compensate(ctx)
        }
    }

    return nil
}
```

### 8. Event Correlation
**Problem:** No way to track event causation and correlation across distributed systems.

**Solution:**
```go
type CorrelatedEvent interface {
    Event
    CorrelationID() string
    CausationID() string
    SetCorrelationID(id string)
    SetCausationID(id string)
}

type BaseEvent struct {
    ID           string    `json:"id"`
    Type         string    `json:"type"`
    Time         time.Time `json:"time"`
    Aggregate    string    `json:"aggregate_id"`
    Version      int       `json:"version"`
    Correlation  string    `json:"correlation_id"`
    Causation    string    `json:"causation_id"`
}
```

## Observability

### 9. Metrics and Tracing
**Problem:** No built-in observability for monitoring system health.

**Solution:**
```go
type MetricsCollector interface {
    RecordCommandDuration(cmd string, duration time.Duration)
    RecordEventEmitted(eventType string)
    RecordError(operation string, err error)
}

type TracingMiddleware struct {
    tracer trace.Tracer
}

func (t *TracingMiddleware) Wrap(next HandlerFunc) HandlerFunc {
    return func(ctx context.Context, cmd Command) (any, error) {
        ctx, span := t.tracer.Start(ctx, fmt.Sprintf("Handle%T", cmd))
        defer span.End()

        result, err := next(ctx, cmd)
        if err != nil {
            span.RecordError(err)
            span.SetStatus(codes.Error, err.Error())
        }

        return result, err
    }
}
```

### 10. Event Replay and Debugging
**Problem:** No way to replay events for debugging or system recovery.

**Solution:**
```go
type EventReplayer struct {
    store EventStore
    app   *App
}

func (r *EventReplayer) Replay(ctx context.Context, from, to time.Time, filter func(Event) bool) error {
    events, err := r.store.LoadTimeRange(ctx, from, to)
    if err != nil {
        return err
    }

    for _, event := range events {
        if filter != nil && !filter(event) {
            continue
        }

        if err := r.app.HandleEvent(ctx, event); err != nil {
            log.Printf("Failed to replay event %s: %v", event.EventID(), err)
        }
    }

    return nil
}
```

## API Improvements

### 11. Generic Handler Registration
**Problem:** Reflection-based handler registration is not type-safe and requires runtime type assertions.

**Solution:**
```go
func RegisterHandler[T Command](app *App, handler func(context.Context, T) (any, error)) {
    app.handlers[reflect.TypeOf((*T)(nil)).Elem()] = func(ctx context.Context, cmd Command) (any, error) {
        typedCmd, ok := cmd.(T)
        if !ok {
            return nil, fmt.Errorf("invalid command type: expected %T, got %T", (*T)(nil), cmd)
        }
        return handler(ctx, typedCmd)
    }
}
```

### 12. Command Metadata
**Problem:** Commands have no metadata, making it difficult to add common concerns.

**Solution:**
```go
type Command interface {
    CommandID() string
    CommandType() string
    Timestamp() time.Time
    UserID() string
}

type BaseCommand struct {
    ID     string    `json:"id"`
    Type   string    `json:"type"`
    Time   time.Time `json:"time"`
    User   string    `json:"user_id"`
}
```

## Testing Improvements

### 13. Test Utilities
**Problem:** No built-in testing utilities for common scenarios.

**Solution:**
```go
type TestEventStore struct {
    events []Event
    mu     sync.Mutex
}

type TestApp struct {
    *App
    RecordedCommands []Command
    RecordedEvents   []Event
}

func NewTestApp() *TestApp {
    return &TestApp{
        App:              NewApp(),
        RecordedCommands: make([]Command, 0),
        RecordedEvents:   make([]Event, 0),
    }
}
```

## Implementation Priority

### High Priority (Security & Stability)
1. Thread safety fixes
2. Error handling improvements
3. Context propagation

### Medium Priority (Functionality)
4. Event store abstraction
5. Middleware support
6. Correlation/causation tracking

### Low Priority (Nice-to-have)
7. Performance optimizations
8. Advanced patterns (Saga, CQRS views)
9. Observability features

## Breaking Changes

The following improvements would require breaking changes:
- Adding methods to the Command interface
- Changing Event interface to include correlation
- Modifying HandlerFunc signature to return errors properly

Consider implementing these in a v2 release with a migration guide.

## Migration Path

For users upgrading from v1 to v2:
1. Provide compatibility layer for old interfaces
2. Add deprecation warnings in v1.x releases
3. Offer automated migration tool for common patterns
4. Maintain v1 branch for critical security fixes

## Benchmark Targets

After implementing improvements, aim for:
- Event emission: < 1μs per event
- Command handling: < 10μs overhead
- Concurrent operations: Support 10,000+ ops/sec
- Memory usage: < 1KB per event in flight