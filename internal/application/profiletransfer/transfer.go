// Package profiletransfer implements explicit, offline profile backup and restore.
package profiletransfer

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"wireglow/internal/domain/profile"
	"wireglow/internal/freeturn"
)

const MaxBytes = 16 * 1024 * 1024
const MaxProfiles = 1000

var (
	ErrInvalid = errors.New("invalid profile archive")
	ErrExpired = errors.New("profile preview expired")
	ErrChanged = errors.New("profile changed after preview")
)

type Archive struct {
	Format   string            `json:"format"`
	Version  int               `json:"version"`
	Profiles []profile.Profile `json:"profiles"`
}

type Store interface {
	List(context.Context) ([]profile.Summary, error)
	Get(context.Context, string) (profile.Profile, error)
	// Lookup distinguishes an absent profile from an unreadable one.
	Lookup(context.Context, string) (*profile.Profile, error)
	SaveTransferred(context.Context, profile.Profile, *profile.Profile, bool) (profile.Summary, error)
}
type Files interface {
	Read(string) ([]byte, error)
	Write(string, []byte) error
}
type Entry struct {
	Profile  profile.Summary
	Existing *profile.Summary
}
type Preview struct {
	Token   string
	Entries []Entry
}
type Choice struct {
	Index  int
	Action string
}
type Result struct {
	Index  int
	Status string
}
type pending struct {
	token    string
	expires  time.Time
	profiles []profile.Profile
	existing []*profile.Profile
}
type Service struct {
	Store   Store
	Files   Files
	mu      sync.Mutex
	pending *pending
}

func validID(id string) bool {
	b, err := hex.DecodeString(id)
	return err == nil && len(b) == 16 && hex.EncodeToString(b) == id
}

// Decode validates the whole archive before any store mutation. Never expose JSON errors.
func Decode(data []byte) (*Archive, error) {
	if len(data) == 0 || len(data) > MaxBytes {
		return nil, ErrInvalid
	}
	var a Archive
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(&a) != nil || d.Decode(new(any)) != io.EOF || a.Format != "wireglow-profiles" || a.Version != 1 || len(a.Profiles) == 0 || len(a.Profiles) > MaxProfiles {
		return nil, ErrInvalid
	}
	seen := make(map[string]bool)
	for _, p := range a.Profiles {
		encoded, err := json.Marshal(p)
		if err != nil || len(encoded) > 256*1024 || !validID(p.ID) || seen[p.ID] || !validConnection(p.Connection) {
			return nil, ErrInvalid
		}
		seen[p.ID] = true
	}
	return &a, nil
}
func validConnection(c freeturn.Connection) bool {
	if len([]rune(c.Name)) > 80 || strings.TrimSpace(c.Peer) == "" || (c.Provider != freeturn.ProviderVK && c.Provider != freeturn.ProviderDirect) || c.Streams < 0 || c.Streams > 1024 || c.StreamsPerCredential < 0 || c.StreamsPerCredential > 1024 {
		return false
	}
	if c.Provider == freeturn.ProviderVK {
		if c.VK == nil || len(c.VK.CallLinks) == 0 {
			return false
		}
		for _, link := range c.VK.CallLinks {
			if strings.TrimSpace(link) == "" || strings.Contains(link, ",") {
				return false
			}
		}
		if _, _, err := freeturn.SplitTURNAddress(c.VK.TurnAddress); err != nil {
			return false
		}
	}
	return true
}

func (s *Service) Export(ctx context.Context, ids []string, path string) error {
	if len(ids) == 0 || len(ids) > MaxProfiles {
		return ErrInvalid
	}
	a := Archive{Format: "wireglow-profiles", Version: 1}
	seen := make(map[string]bool)
	for _, id := range ids {
		if !validID(id) || seen[id] {
			return ErrInvalid
		}
		seen[id] = true
		p, err := s.Store.Get(ctx, id)
		if err != nil {
			return err
		}
		a.Profiles = append(a.Profiles, p)
	}
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}
	// Apply the same limits on both directions, so every export is importable.
	data = append(data, '\n')
	if _, err = Decode(data); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	return s.Files.Write(path, data)
}
func (s *Service) Preview(ctx context.Context, path string) (Preview, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Opening another file invalidates old secrets and prevents unbounded retention.
	s.pending = nil
	data, err := s.Files.Read(path)
	if err != nil {
		return Preview{}, err
	}
	a, err := Decode(data)
	if err != nil {
		return Preview{}, err
	}
	token := make([]byte, 32)
	if _, err = rand.Read(token); err != nil {
		return Preview{}, err
	}
	p := &pending{token: hex.EncodeToString(token), expires: time.Now().Add(10 * time.Minute), profiles: a.Profiles}
	out := Preview{Token: p.token}
	for _, incoming := range a.Profiles {
		existing, err := s.Store.Lookup(ctx, incoming.ID)
		if err != nil {
			return Preview{}, err
		}
		entry := Entry{Profile: incoming.Summary()}
		if existing != nil {
			summary := existing.Summary()
			entry.Existing = &summary
		}
		p.existing = append(p.existing, existing)
		out.Entries = append(out.Entries, entry)
	}
	s.pending = p
	return out, nil
}
func (s *Service) Import(ctx context.Context, token string, choices []Choice) ([]Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.pending
	if p == nil || p.token != token || time.Now().After(p.expires) {
		return nil, ErrExpired
	}
	if len(choices) == 0 || len(choices) > len(p.profiles) {
		return nil, ErrInvalid
	}
	seen := make(map[int]bool)
	for _, c := range choices {
		if c.Index < 0 || c.Index >= len(p.profiles) || seen[c.Index] || (c.Action != "skip" && c.Action != "copy" && c.Action != "replace" && c.Action != "add") || (c.Action == "add" && p.existing[c.Index] != nil) || (c.Action == "replace" && p.existing[c.Index] == nil) {
			return nil, ErrInvalid
		}
		seen[c.Index] = true
	}
	// Single-use preview: retrying an uncertain RPC cannot silently create extra copies.
	s.pending = nil
	results := make([]Result, 0, len(choices))
	for _, c := range choices {
		result := Result{Index: c.Index, Status: "skipped"}
		if c.Action != "skip" {
			_, err := s.Store.SaveTransferred(ctx, p.profiles[c.Index], p.existing[c.Index], c.Action == "copy")
			switch {
			case err == nil:
				result.Status = "imported"
			case errors.Is(err, ErrChanged):
				result.Status = "changed"
			default:
				result.Status = "failed"
			}
		}
		results = append(results, result)
	}
	return results, nil
}
