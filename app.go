package gocmdevt

import (
	"context"
	"fmt"
	"reflect"
)

type Command interface{}
type HandlerFunc func(ctx context.Context, cmd Command) (any, error)

type Module interface {
	Handlers() map[reflect.Type]HandlerFunc
}

type App struct {
	handlers map[reflect.Type]HandlerFunc
}

func NewApp(modules ...Module) *App {
	app := &App{
		handlers: map[reflect.Type]HandlerFunc{},
	}
	for _, m := range modules {
		for typ, h := range m.Handlers() {
			app.handlers[typ] = h
		}
	}
	return app
}

func (a *App) Handle(ctx context.Context, cmd Command) (any, error) {
	handler, ok := a.handlers[reflect.TypeOf(cmd)]
	if !ok {
		return nil, fmt.Errorf("no handler for command type: %T", cmd)
	}
	return handler(ctx, cmd)
}
