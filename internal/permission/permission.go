// Package permission manages access control for tools and other resources.
package permission

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Decision is the result of a permission check.
type Decision int

const (
	DecisionDeny Decision = iota
	DecisionAllow
	DecisionAsk
)

func (d Decision) String() string {
	switch d {
	case DecisionDeny:
		return "deny"
	case DecisionAllow:
		return "allow"
	case DecisionAsk:
		return "ask"
	default:
		return "unknown"
	}
}

// Level indicates the duration of a permission grant.
type Level int

const (
	LevelOnce Level = iota
	LevelSession
	LevelForever
)

func (l Level) String() string {
	switch l {
	case LevelOnce:
		return "once"
	case LevelSession:
		return "session"
	case LevelForever:
		return "forever"
	default:
		return "unknown"
	}
}

// ParseLevel converts a string to a Level.
func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(s) {
	case "once":
		return LevelOnce, nil
	case "session":
		return LevelSession, nil
	case "forever":
		return LevelForever, nil
	default:
		return LevelOnce, fmt.Errorf("unknown permission level: %s", s)
	}
}

// ParseDecision converts a string to a Decision.
func ParseDecision(s string) (Decision, error) {
	switch strings.ToLower(s) {
	case "deny":
		return DecisionDeny, nil
	case "allow":
		return DecisionAllow, nil
	case "ask":
		return DecisionAsk, nil
	default:
		return DecisionAsk, fmt.Errorf("unknown decision: %s", s)
	}
}

// Action represents something that requires permission.
type Action struct {
	Scope string // e.g., "mcp", "model", "file"
	Name  string // e.g., "brave-search_search", "llama-3.2-3b"
}

// String returns the action in "scope:name" format.
func (a Action) String() string {
	return a.Scope + ":" + a.Name
}

// ParseAction parses a "scope:name" string into an Action.
func ParseAction(s string) (Action, error) {
	idx := strings.Index(s, ":")
	if idx < 0 {
		return Action{}, fmt.Errorf("invalid action format (expected scope:name): %s", s)
	}
	return Action{
		Scope: s[:idx],
		Name:  s[idx+1:],
	}, nil
}

// Rule associates a pattern with a decision and grant level.
type Rule struct {
	Scope    string   `yaml:"scope"`
	Name     string   `yaml:"name"`
	Decision Decision `yaml:"decision"`
	Level    Level    `yaml:"level"`
}

// Match checks if an action matches this rule.
func (r Rule) Match(a Action) bool {
	return matchGlob(r.Scope, a.Scope) && matchGlob(r.Name, a.Name)
}

// Grant records a permission grant.
type Grant struct {
	Action    Action `json:"action"`
	Level     Level  `json:"level"`
	SessionID string `json:"session_id,omitempty"`
}

// matchGlob checks if a string matches a glob pattern.
// Supports * and ? wildcards.
func matchGlob(pattern, value string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	matched, err := filepath.Match(pattern, value)
	return err == nil && matched
}
