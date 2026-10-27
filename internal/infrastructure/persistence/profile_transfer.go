package persistence

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"wireglow/internal/application/profiletransfer"
	"wireglow/internal/domain/profile"
)

func (s *Profiles) Lookup(ctx context.Context, id string) (*profile.Profile, error) {
	p, err := s.Get(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// SaveTransferred rechecks the exact previewed contents under the store lock.
// Each file is atomic; failures in a batch are reported individually.
func (s *Profiles) SaveTransferred(ctx context.Context, incoming profile.Profile, expected *profile.Profile, copyProfile bool) (profile.Summary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, err := s.get(ctx, incoming.ID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return profile.Summary{}, err
	}
	if (expected == nil && err == nil) || (expected != nil && (err != nil || !reflect.DeepEqual(current, *expected))) {
		return profile.Summary{}, profiletransfer.ErrChanged
	}
	if copyProfile {
		id := make([]byte, 16)
		if _, err = rand.Read(id); err != nil {
			return profile.Summary{}, err
		}
		incoming.ID = hex.EncodeToString(id)
		// Do not overwrite even in the event of an ID collision.
		if _, err = s.get(ctx, incoming.ID); !errors.Is(err, ErrNotFound) {
			return profile.Summary{}, profiletransfer.ErrChanged
		}
		name := []rune(incoming.Connection.Name)
		if len(name) > 73 {
			name = name[:73]
		}
		incoming.Connection.Name = string(name) + " (copy)"
	}
	if err = ctx.Err(); err != nil {
		return profile.Summary{}, err
	}
	if err = os.MkdirAll(s.Dir, 0700); err != nil {
		return profile.Summary{}, err
	}
	if err = WriteJSON(filepath.Join(s.Dir, incoming.ID+".json"), incoming); err != nil {
		return profile.Summary{}, err
	}
	return incoming.Summary(), nil
}

// ProfileFiles keeps raw archive data in the Go process. Paths come only from an explicit picker.
type ProfileFiles struct{ ProfilesDir string }

func (ProfileFiles) Read(path string) ([]byte, error) {
	if !filepath.IsAbs(path) {
		return nil, profiletransfer.ErrInvalid
	}
	f, err := openReadFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !stat.Mode().IsRegular() || stat.Size() > profiletransfer.MaxBytes {
		return nil, profiletransfer.ErrInvalid
	}
	data, err := io.ReadAll(io.LimitReader(f, profiletransfer.MaxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > profiletransfer.MaxBytes {
		return nil, profiletransfer.ErrInvalid
	}
	return data, nil
}
func (f ProfileFiles) Write(path string, data []byte) error {
	if !filepath.IsAbs(path) {
		return profiletransfer.ErrInvalid
	}
	// An export must never replace the controller's profile store.
	dir, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		return err
	}
	store, err := filepath.Abs(f.ProfilesDir)
	if err != nil {
		return err
	}
	if resolved, e := filepath.EvalSymlinks(store); e == nil {
		store = resolved
	}
	relative, err := filepath.Rel(store, filepath.Join(dir, filepath.Base(path)))
	if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return profiletransfer.ErrInvalid
	}
	return WriteAtomic(path, data)
}
