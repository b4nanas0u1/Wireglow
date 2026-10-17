package ports

import (
	"context"
	"time"
	"wireglow/internal/domain/profile"
)

// Process сообщает о завершении закрытием Done; результат можно читать многократно.
type Process interface {
	Done() <-chan struct{}
	Err() error
	Stop(context.Context) error
}

type Command struct {
	Path string
	Args []string
	Dir  string
}

type Runner interface {
	Start(context.Context, Command) (Process, error)
}

// Plan содержит данные одной попытки; секреты не входят в snapshot контроллера.
type Plan struct {
	Relay          Command
	Engine         Command
	MixedAddress   string
	Ready          func(context.Context) error
	Probe          func(context.Context) error
	MeasureLatency func(context.Context) (time.Duration, error)
	Close          func() error
}

type Preparer interface {
	Prepare(context.Context, profile.Profile) (*Plan, error)
}

// Network сохраняет журнал до изменений и восстанавливает только собственные значения.
type Network interface {
	Recover(context.Context) error
	Apply(context.Context, string) error
}
