package permission

import (
	"context"
	"fmt"
	"sync"
)

// AskFunc is called when a permission check results in Ask and no rule matches.
// The function should prompt the user and return the granted Level (or an error to deny).
type AskFunc func(ctx context.Context, action Action) (Level, error)

// Manager controls access to resources based on configured rules and grants.
type Manager struct {
	mu        sync.RWMutex
	rules     []Rule
	store     Store
	sessionID string
	askFn     AskFunc
}

// NewManager creates a permission manager.
func NewManager(sessionID string, store Store) *Manager {
	return &Manager{
		sessionID: sessionID,
		store:     store,
	}
}

// SetRules replaces the configured permission rules.
func (m *Manager) SetRules(rules []Rule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rules = rules
}

// SetAskFunc sets the callback for when user approval is needed.
func (m *Manager) SetAskFunc(fn AskFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.askFn = fn
}

// Check determines whether an action is allowed and returns the effective level.
// If the action is denied, an error is returned.
// If the action requires Ask, the AskFunc callback is invoked.
func (m *Manager) Check(ctx context.Context, action Action) (Level, error) {
	m.mu.RLock()
	askFn := m.askFn
	m.mu.RUnlock()

	// Check persistent grants first (LevelForever)
	if m.store != nil {
		grants, err := m.store.List(ctx)
		if err == nil {
			for _, g := range grants {
				if g.Action.Scope == action.Scope && g.Action.Name == action.Name {
					if g.Level == LevelForever {
						return LevelForever, nil
					}
					if g.Level == LevelSession && g.SessionID == m.sessionID {
						return LevelSession, nil
					}
				}
			}
		}
	}

	// Check configured rules
	m.mu.RLock()
	rules := m.rules
	m.mu.RUnlock()

	var matchedAsk bool
	for _, rule := range rules {
		if rule.Match(action) {
			switch rule.Decision {
			case DecisionAllow:
				return rule.Level, nil
			case DecisionDeny:
				return rule.Level, fmt.Errorf("permission denied: %s (rule: %s:%s)", action, rule.Scope, rule.Name)
			case DecisionAsk:
				matchedAsk = true
			}
		}
	}

	// If any rule matched with ask, or no rules matched, prompt user
	if matchedAsk || len(rules) == 0 {
		if askFn != nil {
			level, err := askFn(ctx, action)
			if err != nil {
				return LevelOnce, fmt.Errorf("permission denied: %s: %w", action, err)
			}
			// Persist the grant
			if level >= LevelSession && m.store != nil {
				grant := Grant{
					Action:    action,
					Level:     level,
					SessionID: m.sessionID,
				}
				m.store.Save(ctx, grant)
			}
			return level, nil
		}
		return LevelOnce, fmt.Errorf("permission denied: %s (no ask handler)", action)
	}

	return LevelOnce, fmt.Errorf("permission denied: %s (no matching rule)", action)
}

// Revoke removes all grants for a given action.
func (m *Manager) Revoke(ctx context.Context, action Action) error {
	if m.store == nil {
		return nil
	}
	return m.store.Delete(ctx, action)
}

// RevokeSession removes all session-level grants.
func (m *Manager) RevokeSession(ctx context.Context) error {
	if m.store == nil {
		return nil
	}
	grants, err := m.store.List(ctx)
	if err != nil {
		return err
	}
	for _, g := range grants {
		if g.SessionID == m.sessionID {
			m.store.Delete(ctx, g.Action)
		}
	}
	return nil
}
