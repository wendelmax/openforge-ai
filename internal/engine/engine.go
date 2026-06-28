// Package engine manages chat sessions and context.
package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/openforge-ai/openforge/runtime"
)

// Session represents a chat session with messages and a model assignment.
type Session struct {
	ID        string    `json:"id"`
	ModelID   string    `json:"model_id"`
	Messages  []runtime.Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

// Engine manages chat sessions and coordinates with the AI runtime.
type Engine struct {
	mu       sync.RWMutex
	runtime  runtime.Runtime
	sessions map[string]*Session
}

// New creates a new Engine with the given runtime.
func New(rt runtime.Runtime) *Engine {
	return &Engine{
		runtime:  rt,
		sessions: make(map[string]*Session),
	}
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
	e.sessions[session.ID] = session
	return session, nil
}

// GetSession retrieves a session by ID, returning an error if not found.
func (e *Engine) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	session, ok := e.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return session, nil
}

// DeleteSession removes a session by ID, returning an error if not found.
func (e *Engine) DeleteSession(ctx context.Context, sessionID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.sessions[sessionID]; !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	delete(e.sessions, sessionID)
	return nil
}

// ClearSessions removes all sessions from the engine.
func (e *Engine) ClearSessions(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.sessions = make(map[string]*Session)
	return nil
}

// AddMessage appends a message to a session and trims old messages if the token budget is exceeded.
func (e *Engine) AddMessage(ctx context.Context, sessionID string, msg runtime.Message) (*Session, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	session, ok := e.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
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
	return e.runtime.Close(ctx)
}
