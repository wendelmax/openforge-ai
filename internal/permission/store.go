package permission

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Store persists permission grants.
type Store interface {
	Save(ctx context.Context, grant Grant) error
	Delete(ctx context.Context, action Action) error
	List(ctx context.Context) ([]Grant, error)
	Close(ctx context.Context) error
}

// MemoryStore stores grants in memory (ephemeral).
type MemoryStore struct {
	mu     sync.RWMutex
	grants []Grant
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) Save(_ context.Context, grant Grant) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Replace existing grant for the same action
	for i, g := range s.grants {
		if g.Action.Scope == grant.Action.Scope && g.Action.Name == grant.Action.Name {
			s.grants[i] = grant
			return nil
		}
	}
	s.grants = append(s.grants, grant)
	return nil
}

func (s *MemoryStore) Delete(_ context.Context, action Action) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, g := range s.grants {
		if g.Action.Scope == action.Scope && g.Action.Name == action.Name {
			s.grants = append(s.grants[:i], s.grants[i+1:]...)
			return nil
		}
	}
	return nil
}

func (s *MemoryStore) List(_ context.Context) ([]Grant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Grant, len(s.grants))
	copy(result, s.grants)
	return result, nil
}

func (s *MemoryStore) Close(_ context.Context) error {
	return nil
}

// FileStore persists grants as a JSON file.
type FileStore struct {
	mu     sync.Mutex
	path   string
	grants []Grant
}

func NewFileStore(path string) (*FileStore, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create grants dir: %w", err)
	}
	s := &FileStore{path: path}
	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, &s.grants)
	}
	if s.grants == nil {
		s.grants = make([]Grant, 0)
	}
	return s, nil
}

func (s *FileStore) Save(_ context.Context, grant Grant) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, g := range s.grants {
		if g.Action.Scope == grant.Action.Scope && g.Action.Name == grant.Action.Name {
			s.grants[i] = grant
			return s.write()
		}
	}
	s.grants = append(s.grants, grant)
	return s.write()
}

func (s *FileStore) Delete(_ context.Context, action Action) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, g := range s.grants {
		if g.Action.Scope == action.Scope && g.Action.Name == action.Name {
			s.grants = append(s.grants[:i], s.grants[i+1:]...)
			return s.write()
		}
	}
	return nil
}

func (s *FileStore) List(_ context.Context) ([]Grant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Grant, len(s.grants))
	copy(result, s.grants)
	return result, nil
}

func (s *FileStore) Close(_ context.Context) error {
	return nil
}

func (s *FileStore) write() error {
	data, err := json.MarshalIndent(s.grants, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal grants: %w", err)
	}
	return os.WriteFile(s.path, data, 0644)
}
