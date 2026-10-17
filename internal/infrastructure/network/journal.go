// Package network реализует восстановление настроек по журналу принадлежности.
package network

import (
	"context"
	"errors"
	"os"
	"sync"

	"wireglow/internal/infrastructure/persistence"
)

// Settings читает и записывает независимые настройки ОС в каноническом виде.
type Settings interface {
	Desired(string) ([]Change, error)
	Read(context.Context, string) (string, error)
	Write(context.Context, string, string) error
}

type Change struct {
	Key      string `json:"key"`
	Before   string `json:"before"`
	After    string `json:"after"`
	Restored bool   `json:"restored"`
	Conflict bool   `json:"conflict"`
}
type Journal struct {
	Version int      `json:"version"`
	Changes []Change `json:"changes"`
}

var ErrConflict = errors.New("network: settings changed externally")

type Manager struct {
	Path     string
	Settings Settings
	mu       sync.Mutex
}

func (m *Manager) Apply(ctx context.Context, address string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.recover(ctx); err != nil {
		return err
	}
	changes, err := m.Settings.Desired(address)
	if err != nil {
		return err
	}
	j := Journal{Version: 1}
	for _, change := range changes {
		if err := ctx.Err(); err != nil {
			return err
		}
		before, err := m.Settings.Read(ctx, change.Key)
		if err != nil {
			return err
		}
		change.Before = before
		// Намерение записано до изменения ОС, включая падение внутри Write.
		j.Changes = append(j.Changes, change)
		if err := persistence.WriteJSON(m.Path, j); err != nil {
			return err
		}
		if err := m.Settings.Write(ctx, change.Key, change.After); err != nil {
			return err
		}
		actual, err := m.Settings.Read(ctx, change.Key)
		if err != nil {
			return err
		}
		if actual != change.After {
			return errors.New("network: setting verification failed")
		}
	}
	return nil
}

func (m *Manager) Recover(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.recover(ctx)
}

func (m *Manager) recover(ctx context.Context) error {
	var j Journal
	if err := persistence.ReadJSON(m.Path, &j); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if j.Version != 1 {
		return errors.New("network: unsupported journal version")
	}
	var result error
	// Восстанавливаем в обратном порядке: режим отключается до изменения адресов.
	for i := len(j.Changes) - 1; i >= 0; i-- {
		change := &j.Changes[i]
		if change.Restored {
			continue
		}
		if change.Conflict {
			result = errors.Join(result, ErrConflict)
			continue
		}
		actual, err := m.Settings.Read(ctx, change.Key)
		if err != nil {
			result = errors.Join(result, err)
			continue
		}
		switch actual {
		case change.Before:
			change.Restored = true
		case change.After:
			if err := m.Settings.Write(ctx, change.Key, change.Before); err != nil {
				result = errors.Join(result, err)
				continue
			}
			actual, err = m.Settings.Read(ctx, change.Key)
			if err != nil || actual != change.Before {
				result = errors.Join(result, errors.New("network: restore verification failed"))
				continue
			}
			change.Restored = true
		default:
			change.Conflict = true
			result = errors.Join(result, ErrConflict)
		}
		if err := persistence.WriteJSON(m.Path, j); err != nil {
			return errors.Join(result, err)
		}
	}
	if result != nil {
		return result
	}
	return os.Remove(m.Path)
}

// Manual не меняет системную настройку; mixed endpoint остаётся доступным приложению.
type Manual struct{}

func (Manual) Apply(context.Context, string) error { return nil }
func (Manual) Recover(context.Context) error       { return nil }
