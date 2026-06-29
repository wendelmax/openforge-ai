package engine

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/openforge-ai/openforge/internal/skill"
	"github.com/openforge-ai/openforge/runtime"
)

// Session represents a chat session with messages and a model assignment.
type Session struct {
	ID        string             `json:"id"`
	ModelID   string             `json:"model_id"`
	Messages  []runtime.Message  `json:"messages"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	MaxTokens int                `json:"max_tokens,omitempty"`
}

// Engine manages chat sessions and coordinates with the AI runtime.
type Engine struct {
	mu       sync.RWMutex
	runtime  runtime.Runtime
	store    SessionStore
	executor *skill.Executor
}

// New creates a new Engine with SQLite session persistence, falling back to memory.
func New(rt runtime.Runtime) *Engine {
	store, err := NewSQLiteSessionStore("./data/sessions.db")
	if err != nil {
		log.Printf("sqlite store init failed, using memory: %v", err)
		store = nil
	}
	var s SessionStore
	if store != nil {
		s = store
	} else {
		s = NewMemorySessionStore()
	}
	return &Engine{
		runtime:  rt,
		store:    s,
		executor: skill.NewExecutor(rt),
	}
}

// NewWithStore creates a new Engine with the given session store.
func NewWithStore(rt runtime.Runtime, store SessionStore) *Engine {
	return &Engine{
		runtime:  rt,
		store:    store,
		executor: skill.NewExecutor(rt),
	}
}

// Store returns the session store used by the engine.
func (e *Engine) Store() SessionStore {
	return e.store
}

// SkillExecutor returns the skill executor.
func (e *Engine) SkillExecutor() *skill.Executor {
	return e.executor
}

// Runtime returns the underlying runtime used by the engine.
func (e *Engine) Runtime() runtime.Runtime {
	return e.runtime
}

// CreateSession creates a new session for the given model and returns it.
func (e *Engine) CreateSession(ctx context.Context, modelID string) (*Session, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	session := &Session{
		ID:        uuid.New().String(),
		ModelID:   modelID,
		Messages:  make([]runtime.Message, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := e.store.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return session, nil
}

// GetSession retrieves a session by ID, returning an error if not found.
func (e *Engine) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	return e.store.Get(ctx, sessionID)
}

// DeleteSession removes a session by ID, returning an error if not found.
func (e *Engine) DeleteSession(ctx context.Context, sessionID string) error {
	return e.store.Delete(ctx, sessionID)
}

// ClearSessions removes all sessions from the engine.
func (e *Engine) ClearSessions(ctx context.Context) error {
	return e.store.Close(ctx)
}

// AddMessage appends a message to a session and trims old messages if the token budget is exceeded.
func (e *Engine) AddMessage(ctx context.Context, sessionID string, msg runtime.Message) (*Session, error) {
	session, err := e.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	session.Messages = append(session.Messages, msg)
	session.UpdatedAt = time.Now()

	maxTokens := session.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	tokenCount := 0
	for i := len(session.Messages) - 1; i >= 0; i-- {
		tokenCount += len(session.Messages[i].Content) / 4
		if tokenCount > maxTokens {
			session.Messages = session.Messages[i+1:]
			break
		}
	}

	if err := e.store.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}
	return session, nil
}

// BuildPrompt constructs a prompt string from all messages in a session.
func (e *Engine) BuildPrompt(ctx context.Context, sessionID string) (string, error) {
	session, err := e.GetSession(ctx, sessionID)
	if err != nil {
		return "", err
	}

	var prompt string
	for _, msg := range session.Messages {
		prompt += msg.Role + ": " + msg.Content + "\n"
	}
	return prompt, nil
}

// Close shuts down the engine and its underlying runtime.
func (e *Engine) Close(ctx context.Context) error {
	e.store.Close(ctx)
	return e.runtime.Close(ctx)
}
