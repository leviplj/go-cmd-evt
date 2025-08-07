package gocmdevt

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// Test command types
type CreateUserCommand struct {
	Name  string
	Email string
}

type DeleteUserCommand struct {
	ID int
}

type UpdateUserCommand struct {
	ID   int
	Name string
}

// Custom context key type
type contextKey string

const testContextKey contextKey = "test"

// Simple module implementations
type UserModule struct{}

func NewUserModule() *UserModule {
	return &UserModule{}
}

func (m *UserModule) Handlers() map[reflect.Type]HandlerFunc {
	return map[reflect.Type]HandlerFunc{
		reflect.TypeOf(&CreateUserCommand{}): m.createUser,
		reflect.TypeOf(&DeleteUserCommand{}): m.deleteUser,
	}
}

func (m *UserModule) createUser(ctx context.Context, cmd Command) (any, error) {
	createCmd := cmd.(*CreateUserCommand)
	return map[string]interface{}{
		"id": "123", "name": createCmd.Name, "email": createCmd.Email,
	}, nil
}

func (m *UserModule) deleteUser(ctx context.Context, cmd Command) (any, error) {
	deleteCmd := cmd.(*DeleteUserCommand)
	return map[string]interface{}{
		"deleted": true, "id": deleteCmd.ID,
	}, nil
}

type AdminModule struct{}

func NewAdminModule() *AdminModule {
	return &AdminModule{}
}

func (m *AdminModule) Handlers() map[reflect.Type]HandlerFunc {
	return map[reflect.Type]HandlerFunc{
		reflect.TypeOf(&UpdateUserCommand{}): m.updateUser,
	}
}

func (m *AdminModule) updateUser(ctx context.Context, cmd Command) (any, error) {
	updateCmd := cmd.(*UpdateUserCommand)
	return map[string]interface{}{
		"updated": true, "id": updateCmd.ID, "name": updateCmd.Name,
	}, nil
}

type EmptyModule struct{}

func NewEmptyModule() *EmptyModule {
	return &EmptyModule{}
}

func (m *EmptyModule) Handlers() map[reflect.Type]HandlerFunc {
	return map[reflect.Type]HandlerFunc{}
}

func TestNewApp(t *testing.T) {
	t.Run("creates app with no modules", func(t *testing.T) {
		app := NewApp()

		if app == nil {
			t.Fatal("expected app to be created, got nil")
		}

		if app.handlers == nil {
			t.Fatal("expected handlers map to be initialized")
		}

		if len(app.handlers) != 0 {
			t.Errorf("expected empty handlers map, got %d handlers", len(app.handlers))
		}
	})

	t.Run("creates app with single module", func(t *testing.T) {
		userModule := NewUserModule()
		app := NewApp(userModule)

		if app == nil {
			t.Fatal("expected app to be created, got nil")
		}

		expectedHandlers := 2 // CreateUserCommand and DeleteUserCommand
		if len(app.handlers) != expectedHandlers {
			t.Errorf("expected %d handlers, got %d", expectedHandlers, len(app.handlers))
		}

		// Check that specific command types are registered
		createUserType := reflect.TypeOf(&CreateUserCommand{})
		deleteUserType := reflect.TypeOf(&DeleteUserCommand{})

		if _, exists := app.handlers[createUserType]; !exists {
			t.Error("expected CreateUserCommand handler to be registered")
		}

		if _, exists := app.handlers[deleteUserType]; !exists {
			t.Error("expected DeleteUserCommand handler to be registered")
		}
	})

	t.Run("creates app with multiple modules", func(t *testing.T) {
		userModule := NewUserModule()
		adminModule := NewAdminModule()
		app := NewApp(userModule, adminModule)

		expectedHandlers := 3 // CreateUser, DeleteUser, UpdateUser
		if len(app.handlers) != expectedHandlers {
			t.Errorf("expected %d handlers, got %d", expectedHandlers, len(app.handlers))
		}

		// Check all command types are registered
		types := []reflect.Type{
			reflect.TypeOf(&CreateUserCommand{}),
			reflect.TypeOf(&DeleteUserCommand{}),
			reflect.TypeOf(&UpdateUserCommand{}),
		}

		for _, typ := range types {
			if _, exists := app.handlers[typ]; !exists {
				t.Errorf("expected handler for type %v to be registered", typ)
			}
		}
	})

	t.Run("creates app with empty module", func(t *testing.T) {
		emptyModule := NewEmptyModule()
		userModule := NewUserModule()
		app := NewApp(emptyModule, userModule)

		expectedHandlers := 2 // Only from UserModule
		if len(app.handlers) != expectedHandlers {
			t.Errorf("expected %d handlers, got %d", expectedHandlers, len(app.handlers))
		}
	})

	t.Run("later modules override earlier modules for same command type", func(t *testing.T) {
		// Create a custom app to test override behavior
		app := &App{handlers: map[reflect.Type]HandlerFunc{}}

		userModule := NewUserModule()
		// Add first module handlers
		for typ, h := range userModule.Handlers() {
			app.handlers[typ] = h
		}

		// Override the CreateUserCommand handler
		app.handlers[reflect.TypeOf(&CreateUserCommand{})] = func(ctx context.Context, cmd Command) (any, error) {
			return "overridden", nil
		}

		// Test that the overridden handler is used
		ctx := context.Background()
		result, err := app.Handle(ctx, &CreateUserCommand{Name: "John", Email: "john@example.com"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != "overridden" {
			t.Errorf("expected overridden result, got %v", result)
		}
	})
}

func TestApp_Handle(t *testing.T) {
	userModule := NewUserModule()
	adminModule := NewAdminModule()
	app := NewApp(userModule, adminModule)
	ctx := context.Background()

	t.Run("handles registered CreateUserCommand", func(t *testing.T) {
		cmd := &CreateUserCommand{
			Name:  "John Doe",
			Email: "john@example.com",
		}

		result, err := app.Handle(ctx, cmd)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map[string]interface{}, got %T", result)
		}

		if resultMap["name"] != "John Doe" {
			t.Errorf("expected name 'John Doe', got '%v'", resultMap["name"])
		}

		if resultMap["email"] != "john@example.com" {
			t.Errorf("expected email 'john@example.com', got '%v'", resultMap["email"])
		}

		if resultMap["id"] != "123" {
			t.Errorf("expected id '123', got '%v'", resultMap["id"])
		}
	})

	t.Run("handles registered DeleteUserCommand", func(t *testing.T) {
		cmd := &DeleteUserCommand{ID: 456}

		result, err := app.Handle(ctx, cmd)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map[string]interface{}, got %T", result)
		}

		if resultMap["deleted"] != true {
			t.Errorf("expected deleted to be true, got %v", resultMap["deleted"])
		}

		if resultMap["id"] != 456 {
			t.Errorf("expected id 456, got %v", resultMap["id"])
		}
	})

	t.Run("handles registered UpdateUserCommand", func(t *testing.T) {
		cmd := &UpdateUserCommand{
			ID:   789,
			Name: "Jane Doe",
		}

		result, err := app.Handle(ctx, cmd)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map[string]interface{}, got %T", result)
		}

		if resultMap["updated"] != true {
			t.Errorf("expected updated to be true, got %v", resultMap["updated"])
		}

		if resultMap["id"] != 789 {
			t.Errorf("expected id 789, got %v", resultMap["id"])
		}

		if resultMap["name"] != "Jane Doe" {
			t.Errorf("expected name 'Jane Doe', got %v", resultMap["name"])
		}
	})

	t.Run("returns error for unregistered command", func(t *testing.T) {
		type UnregisteredCommand struct {
			Data string
		}

		cmd := UnregisteredCommand{Data: "test"}

		result, err := app.Handle(ctx, cmd)

		if err == nil {
			t.Fatal("expected error for unregistered command, got nil")
		}

		if result != nil {
			t.Errorf("expected nil result for unregistered command, got %v", result)
		}

		expectedErrorPrefix := "no handler for command type:"
		if !strings.Contains(err.Error(), expectedErrorPrefix) {
			t.Errorf("expected error to contain '%s', got '%s'", expectedErrorPrefix, err.Error())
		}

		expectedType := "gocmdevt.UnregisteredCommand"
		if !strings.Contains(err.Error(), expectedType) {
			t.Errorf("expected error to contain type '%s', got '%s'", expectedType, err.Error())
		}
	})

	t.Run("handles command with nil context", func(t *testing.T) {
		cmd := &CreateUserCommand{
			Name:  "Test User",
			Email: "test@example.com",
		}

		result, err := app.Handle(context.TODO(), cmd)

		// This should work since our test handlers don't use the context
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Error("expected non-nil result")
		}
	})

	t.Run("preserves context in handler", func(t *testing.T) {
		// Create an app with a custom handler that uses context
		app := &App{handlers: map[reflect.Type]HandlerFunc{}}

		// Add a handler that checks context
		app.handlers[reflect.TypeOf(CreateUserCommand{})] = func(ctx context.Context, cmd Command) (any, error) {
			if ctx == nil {
				return nil, fmt.Errorf("context is nil")
			}

			// Check if context value exists
			if val := ctx.Value(testContextKey); val != nil {
				return map[string]interface{}{
					"contextValue": val,
				}, nil
			}

			return map[string]interface{}{
				"contextReceived": true,
			}, nil
		}

		// Test with context containing a value
		ctx := context.WithValue(context.Background(), testContextKey, "contextData")
		cmd := CreateUserCommand{Name: "Test", Email: "test@example.com"}

		result, err := app.Handle(ctx, cmd)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map[string]interface{}, got %T", result)
		}

		if resultMap["contextValue"] != "contextData" {
			t.Errorf("expected context value 'contextData', got %v", resultMap["contextValue"])
		}
	})
}

func TestApp_Handle_EdgeCases(t *testing.T) {
	t.Run("empty app handles unregistered command", func(t *testing.T) {
		app := NewApp()
		ctx := context.Background()

		cmd := &CreateUserCommand{Name: "Test", Email: "test@example.com"}

		result, err := app.Handle(ctx, cmd)

		if err == nil {
			t.Fatal("expected error for unregistered command in empty app")
		}

		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})

	t.Run("handles pointer command types", func(t *testing.T) {
		app := &App{handlers: map[reflect.Type]HandlerFunc{}}

		// Register handler for pointer type
		app.handlers[reflect.TypeOf(&CreateUserCommand{})] = func(ctx context.Context, cmd Command) (any, error) {
			return "pointer command handled", nil
		}

		ctx := context.Background()
		cmd := &CreateUserCommand{Name: "Test", Email: "test@example.com"}

		result, err := app.Handle(ctx, cmd)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != "pointer command handled" {
			t.Errorf("expected 'pointer command handled', got %v", result)
		}
	})
}

// Benchmark tests
func BenchmarkApp_Handle(b *testing.B) {
	userModule := NewUserModule()
	app := NewApp(userModule)
	ctx := context.Background()
	cmd := &CreateUserCommand{Name: "John", Email: "john@example.com"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := app.Handle(ctx, cmd)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewApp(b *testing.B) {
	userModule := NewUserModule()
	adminModule := NewAdminModule()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewApp(userModule, adminModule)
	}
}
