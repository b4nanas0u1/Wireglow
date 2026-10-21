package ports

import (
	"context"
	"time"
	"wireglow/internal/domain/profile"
	"wireglow/internal/domain/proxymode"
)

// Process сообщает о завершении закрытием Done; результат можно читать многократно.
type Process interface {
	Done() <-chan struct{}
	Err() error
	Stop(context.Context) error
}

type Command struct {
	CaptureDiagnostics bool
	Path               string
	Args               []string
	Dir                string
	ClearProxyEnv      bool
}

type Runner interface {
	Start(context.Context, Command) (Process, error)
}

// Plan содержит данные одной попытки; секреты не входят в snapshot контроллера.
type Plan struct {
	TunReady       bool
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

// ModePreparer supports bootstrap settings that must be fixed before starting relay.
type ModePreparer interface {
	PrepareForMode(context.Context, profile.Profile, proxymode.Mode) (*Plan, error)
}

// NetworkDiagnostic exposes only adapter-controlled, credential-free failure
// details. Arbitrary process errors must never be copied into snapshots.
type NetworkDiagnostic interface {
	NetworkDiagnostic() string
}

// DiagnosticError is an explicitly sanitized error suitable for snapshots.
// Implementations must not expose profile data, URLs or engine output.
type DiagnosticError interface {
	error
	PublicDiagnostic() string
}

// ProcessDiagnostic returns only fixed credential-free engine error categories.
type ProcessDiagnostic interface {
	ProcessDiagnostic() string
	ResetProcessDiagnostic()
}
