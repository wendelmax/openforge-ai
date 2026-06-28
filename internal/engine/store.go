package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/openforge-ai/openforge/internal/config"
)

// NewStoreFromConfig creates a SessionStore based on the configuration.
func NewStoreFromConfig(cfg *config.Config) (SessionStore, error) {
	switch cfg.Session.Backend {
	case "file":
		return NewFileSessionStore(cfg.Session.Path)
	case "sqlite":
		return NewSQLiteSessionStore(cfg.Session.Path + "/sessions.db")
	default:
		return NewMemorySessionStore(), nil
	}
}

// SessionStore persists chat sessions.
type SessionStore interface {
	Create(ctx context.Context, session *Session) error
	Get(ctx context.Context, id string) (*Session, error)
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*Session, error)
	Close(ctx context.Context) error
}

// MemorySessionStore stores sessions in an in-memory map.
type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions: make(map[string]*Session),
	}
}

func (s *MemorySessionStore) Create(_ context.Context, session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	s.sessions[session.ID] = session
	return nil
}

func (s *MemorySessionStore) Get(_ context.Context, id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return session, nil
}

func (s *MemorySessionStore) Update(_ context.Context, session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[session.ID]; !ok {
		return fmt.Errorf("session not found: %s", session.ID)
	}
	s.sessions[session.ID] = session
	return nil
}

func (s *MemorySessionStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[id]; !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	delete(s.sessions, id)
	return nil
}

func (s *MemorySessionStore) List(_ context.Context) ([]*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		result = append(result, session)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

func (s *MemorySessionStore) Close(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions = make(map[string]*Session)
	return nil
}

// FileSessionStore stores sessions as JSON files (one per session).
type FileSessionStore struct {
	mu   sync.RWMutex
	dir  string
	open map[string]*Session
}

func NewFileSessionStore(dir string) (*FileSessionStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create session dir: %w", err)
	}
	return &FileSessionStore{
		dir:  dir,
		open: make(map[string]*Session),
	}, nil
}

func (s *FileSessionStore) path(id string) string {
	return filepath.Join(s.dir, id+".json")
}

func (s *FileSessionStore) Create(_ context.Context, session *Session) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	if err := os.WriteFile(s.path(session.ID), data, 0644); err != nil {
		return fmt.Errorf("write session file: %w", err)
	}
	s.mu.Lock()
	s.open[session.ID] = session
	s.mu.Unlock()
	return nil
}

func (s *FileSessionStore) Get(_ context.Context, id string) (*Session, error) {
	s.mu.RLock()
	cached, ok := s.open[id]
	s.mu.RUnlock()
	if ok {
		return cached, nil
	}

	data, err := os.ReadFile(s.path(id))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		return nil, fmt.Errorf("read session file: %w", err)
	}
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}
	s.mu.Lock()
	s.open[id] = &session
	s.mu.Unlock()
	return &session, nil
}

func (s *FileSessionStore) Update(_ context.Context, session *Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	if err := os.WriteFile(s.path(session.ID), data, 0644); err != nil {
		return fmt.Errorf("write session file: %w", err)
	}
	s.mu.Lock()
	s.open[session.ID] = session
	s.mu.Unlock()
	return nil
}

func (s *FileSessionStore) Delete(_ context.Context, id string) error {
	if err := os.Remove(s.path(id)); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("session not found: %s", id)
		}
		return fmt.Errorf("remove session file: %w", err)
	}
	s.mu.Lock()
	delete(s.open, id)
	s.mu.Unlock()
	return nil
}

func (s *FileSessionStore) List(ctx context.Context) ([]*Session, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("read session dir: %w", err)
	}
	result := make([]*Session, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		id := entry.Name()[:len(entry.Name())-5]
		session, err := s.Get(ctx, id)
		if err != nil {
			continue
		}
		result = append(result, session)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

func (s *FileSessionStore) Close(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.open = make(map[string]*Session)
	return nil
}
