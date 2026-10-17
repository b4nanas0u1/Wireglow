package engines

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// limitedOutput ограничивает память даже при неожиданном выводе внешнего executable.
type limitedOutput struct {
	mu       sync.Mutex
	data     []byte
	exceeded bool
}

func (w *limitedOutput) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n := len(data)
	space := 256*1024 - len(w.data)
	if len(data) > space {
		w.exceeded = true
		data = data[:space]
	}
	w.data = append(w.data, data...)
	return n, nil
}

func checkRelayContract(ctx context.Context, path string, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "-h")
	cmd.WaitDelay = time.Second
	output := &limitedOutput{}
	cmd.Stdout = output
	cmd.Stderr = output
	err := cmd.Run()
	if ctx.Err() != nil || output.exceeded {
		return errors.New("relay: help contract unavailable")
	}
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			return errors.New("relay: help contract unavailable")
		}
	}
	for index := 0; index < len(args); index++ {
		arg := args[index]
		flag, _, inline := strings.Cut(arg, "=")
		if !strings.Contains(string(output.data), flag) {
			return errors.New("relay: required CLI flag absent")
		}
		if !inline {
			index++
		}
	}
	return nil
}
