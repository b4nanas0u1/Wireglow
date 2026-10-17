package persistence

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"wireglow/internal/domain/profile"
	"wireglow/internal/freeturn"
)

var ErrNotFound = errors.New("profile: not found")

// Profiles используется единственным controller под платформенным instance lock.
type Profiles struct {
	Dir string
	mu  sync.Mutex
}

func (s *Profiles) Import(ctx context.Context, link string) (profile.Summary, error) {
	connection, err := freeturn.Parse(link)
	if err != nil {
		return profile.Summary{}, err
	}
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return profile.Summary{}, err
	}
	p := profile.Profile{ID: hex.EncodeToString(id), Connection: *connection}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return profile.Summary{}, err
	}
	if err := os.MkdirAll(s.Dir, 0700); err != nil {
		return profile.Summary{}, err
	}
	if err := WriteJSON(filepath.Join(s.Dir, p.ID+".json"), p); err != nil {
		return profile.Summary{}, err
	}
	return p.Summary(), nil
}

func validID(id string) bool {
	b, err := hex.DecodeString(id)
	return err == nil && len(b) == 16 && hex.EncodeToString(b) == id
}

func (s *Profiles) Get(ctx context.Context, id string) (profile.Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.get(ctx, id)
}

func (s *Profiles) get(ctx context.Context, id string) (profile.Profile, error) {
	var p profile.Profile
	if err := ctx.Err(); err != nil {
		return p, err
	}
	if !validID(id) {
		return p, ErrNotFound
	}
	if err := ReadJSON(filepath.Join(s.Dir, id+".json"), &p); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return p, ErrNotFound
		}
		return p, err
	}
	if p.ID != id {
		return profile.Profile{}, errors.New("storage: mismatched profile ID")
	}
	return p, nil
}

func (s *Profiles) List(ctx context.Context) ([]profile.Summary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := os.ReadDir(s.Dir)
	if errors.Is(err, os.ErrNotExist) {
		return []profile.Summary{}, nil
	}
	if err != nil {
		return nil, err
	}
	result := []profile.Summary{}
	for _, entry := range entries {
		id := entry.Name()
		if filepath.Ext(id) != ".json" {
			continue
		}
		id = id[:len(id)-5]
		if !validID(id) {
			continue
		}
		p, err := s.get(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, p.Summary())
	}
	slices.SortFunc(result, func(a, b profile.Summary) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	return result, nil
}

func (s *Profiles) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if !validID(id) {
		return ErrNotFound
	}
	err := os.Remove(filepath.Join(s.Dir, id+".json"))
	if errors.Is(err, os.ErrNotExist) {
		return ErrNotFound
	}
	return err
}

// Update сохраняет только после загрузки текущего подключения, сохраняя ID.
func (s *Profiles) Update(ctx context.Context, p profile.Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if !validID(p.ID) {
		return ErrNotFound
	}
	if err := os.MkdirAll(s.Dir, 0700); err != nil {
		return err
	}
	current, err := s.get(ctx, p.ID)
	if err != nil {
		return err
	}
	current.Connection = p.Connection
	return WriteJSON(filepath.Join(s.Dir, p.ID+".json"), current)
}
