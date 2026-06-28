package permission

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionString(t *testing.T) {
	a := Action{Scope: "mcp", Name: "test"}
	assert.Equal(t, "mcp:test", a.String())
}

func TestParseAction(t *testing.T) {
	a, err := ParseAction("mcp:test-tool")
	require.NoError(t, err)
	assert.Equal(t, "mcp", a.Scope)
	assert.Equal(t, "test-tool", a.Name)
}

func TestParseAction_Invalid(t *testing.T) {
	_, err := ParseAction("no-colon")
	assert.Error(t, err)
}

func TestRuleMatch(t *testing.T) {
	rule := Rule{Scope: "mcp", Name: "search", Decision: DecisionAllow, Level: LevelForever}

	assert.True(t, rule.Match(Action{Scope: "mcp", Name: "search"}))
	assert.False(t, rule.Match(Action{Scope: "mcp", Name: "other"}))
	assert.False(t, rule.Match(Action{Scope: "model", Name: "search"}))
}

func TestRuleMatchWildcard(t *testing.T) {
	rule := Rule{Scope: "mcp", Name: "*", Decision: DecisionDeny}

	assert.True(t, rule.Match(Action{Scope: "mcp", Name: "anything"}))
	assert.False(t, rule.Match(Action{Scope: "model", Name: "anything"}))
}

func TestRuleMatchWildcardScope(t *testing.T) {
	rule := Rule{Scope: "*", Name: "test", Decision: DecisionAllow}

	assert.True(t, rule.Match(Action{Scope: "mcp", Name: "test"}))
	assert.True(t, rule.Match(Action{Scope: "model", Name: "test"}))
}

func TestParseLevel(t *testing.T) {
	l, err := ParseLevel("once")
	require.NoError(t, err)
	assert.Equal(t, LevelOnce, l)

	l, err = ParseLevel("session")
	require.NoError(t, err)
	assert.Equal(t, LevelSession, l)

	l, err = ParseLevel("forever")
	require.NoError(t, err)
	assert.Equal(t, LevelForever, l)

	_, err = ParseLevel("invalid")
	assert.Error(t, err)
}

func TestParseDecision(t *testing.T) {
	d, err := ParseDecision("allow")
	require.NoError(t, err)
	assert.Equal(t, DecisionAllow, d)

	d, err = ParseDecision("deny")
	require.NoError(t, err)
	assert.Equal(t, DecisionDeny, d)

	d, err = ParseDecision("ask")
	require.NoError(t, err)
	assert.Equal(t, DecisionAsk, d)

	_, err = ParseDecision("invalid")
	assert.Error(t, err)
}

func TestDecisionString(t *testing.T) {
	assert.Equal(t, "deny", DecisionDeny.String())
	assert.Equal(t, "allow", DecisionAllow.String())
	assert.Equal(t, "ask", DecisionAsk.String())
}

func TestLevelString(t *testing.T) {
	assert.Equal(t, "once", LevelOnce.String())
	assert.Equal(t, "session", LevelSession.String())
	assert.Equal(t, "forever", LevelForever.String())
}

func TestManagerCheck_AllowRule(t *testing.T) {
	m := NewManager("sess-1", NewMemoryStore())
	m.SetRules([]Rule{
		{Scope: "mcp", Name: "search", Decision: DecisionAllow, Level: LevelSession},
	})

	level, err := m.Check(context.Background(), Action{Scope: "mcp", Name: "search"})
	require.NoError(t, err)
	assert.Equal(t, LevelSession, level)
}

func TestManagerCheck_DenyRule(t *testing.T) {
	m := NewManager("sess-1", NewMemoryStore())
	m.SetRules([]Rule{
		{Scope: "mcp", Name: "danger", Decision: DecisionDeny},
	})

	_, err := m.Check(context.Background(), Action{Scope: "mcp", Name: "danger"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "denied")
}

func TestManagerCheck_NoRules_AskFunc(t *testing.T) {
	m := NewManager("sess-1", NewMemoryStore())
	m.SetAskFunc(func(ctx context.Context, a Action) (Level, error) {
		return LevelOnce, nil
	})

	level, err := m.Check(context.Background(), Action{Scope: "mcp", Name: "unknown-tool"})
	require.NoError(t, err)
	assert.Equal(t, LevelOnce, level)
}

func TestManagerCheck_NoRules_AskDenied(t *testing.T) {
	m := NewManager("sess-1", NewMemoryStore())
	m.SetAskFunc(func(ctx context.Context, a Action) (Level, error) {
		return LevelOnce, io.EOF
	})

	_, err := m.Check(context.Background(), Action{Scope: "mcp", Name: "unknown-tool"})
	assert.Error(t, err)
}

func TestManagerCheck_NoRules_NoAskFunc(t *testing.T) {
	m := NewManager("sess-1", NewMemoryStore())

	_, err := m.Check(context.Background(), Action{Scope: "mcp", Name: "tool"})
	assert.Error(t, err)
}

func TestManagerCheck_ForeverGrant(t *testing.T) {
	store := NewMemoryStore()
	m := NewManager("sess-1", store)

	store.Save(context.Background(), Grant{
		Action:    Action{Scope: "mcp", Name: "trusted"},
		Level:     LevelForever,
		SessionID: "",
	})

	level, err := m.Check(context.Background(), Action{Scope: "mcp", Name: "trusted"})
	require.NoError(t, err)
	assert.Equal(t, LevelForever, level)
}

func TestManagerCheck_SessionGrant(t *testing.T) {
	store := NewMemoryStore()
	m := NewManager("sess-1", store)

	store.Save(context.Background(), Grant{
		Action:    Action{Scope: "mcp", Name: "sess-tool"},
		Level:     LevelSession,
		SessionID: "sess-1",
	})

	level, err := m.Check(context.Background(), Action{Scope: "mcp", Name: "sess-tool"})
	require.NoError(t, err)
	assert.Equal(t, LevelSession, level)
}

func TestManagerCheck_SessionGrant_OtherSession(t *testing.T) {
	store := NewMemoryStore()
	m := NewManager("sess-2", store)

	store.Save(context.Background(), Grant{
		Action:    Action{Scope: "mcp", Name: "sess-tool"},
		Level:     LevelSession,
		SessionID: "sess-1",
	})

	_, err := m.Check(context.Background(), Action{Scope: "mcp", Name: "sess-tool"})
	assert.Error(t, err)
}

func TestManagerRevoke(t *testing.T) {
	store := NewMemoryStore()
	store.Save(context.Background(), Grant{
		Action: Action{Scope: "mcp", Name: "gone"},
		Level:  LevelForever,
	})

	m := NewManager("sess-1", store)
	err := m.Revoke(context.Background(), Action{Scope: "mcp", Name: "gone"})
	require.NoError(t, err)

	_, err = m.Check(context.Background(), Action{Scope: "mcp", Name: "gone"})
	assert.Error(t, err)
}

func TestManagerRevokeSession(t *testing.T) {
	store := NewMemoryStore()
	store.Save(context.Background(), Grant{
		Action:    Action{Scope: "mcp", Name: "a"},
		Level:     LevelSession,
		SessionID: "sess-1",
	})
	store.Save(context.Background(), Grant{
		Action:    Action{Scope: "mcp", Name: "b"},
		Level:     LevelForever,
		SessionID: "",
	})

	m := NewManager("sess-1", store)
	m.RevokeSession(context.Background())

	_, err := m.Check(context.Background(), Action{Scope: "mcp", Name: "a"})
	assert.Error(t, err)

	_, err = m.Check(context.Background(), Action{Scope: "mcp", Name: "b"})
	assert.NoError(t, err)
}

func TestMemoryStore_List(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	s.Save(ctx, Grant{Action: Action{Scope: "mcp", Name: "a"}, Level: LevelOnce})
	s.Save(ctx, Grant{Action: Action{Scope: "mcp", Name: "b"}, Level: LevelForever})

	list, err := s.List(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestMemoryStore_Delete(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	s.Save(ctx, Grant{Action: Action{Scope: "mcp", Name: "a"}, Level: LevelOnce})
	s.Delete(ctx, Action{Scope: "mcp", Name: "a"})

	list, _ := s.List(ctx)
	assert.Len(t, list, 0)
}

func TestMemoryStore_Replace(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()

	s.Save(ctx, Grant{Action: Action{Scope: "mcp", Name: "a"}, Level: LevelOnce})
	s.Save(ctx, Grant{Action: Action{Scope: "mcp", Name: "a"}, Level: LevelForever})

	list, _ := s.List(ctx)
	require.Len(t, list, 1)
	assert.Equal(t, LevelForever, list[0].Level)
}

func TestFileStore_Persistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "grants.json")
	ctx := context.Background()

	s1, err := NewFileStore(path)
	require.NoError(t, err)
	s1.Save(ctx, Grant{Action: Action{Scope: "mcp", Name: "persist"}, Level: LevelForever})
	s1.Close(ctx)

	s2, err := NewFileStore(path)
	require.NoError(t, err)
	list, err := s2.List(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "persist", list[0].Action.Name)
	assert.Equal(t, LevelForever, list[0].Level)
}

func TestFileStore_DirCreation(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "deep", "path")
	path := filepath.Join(dir, "grants.json")
	s, err := NewFileStore(path)
	require.NoError(t, err)
	assert.DirExists(t, dir)

	ctx := context.Background()
	s.Save(ctx, Grant{Action: Action{Scope: "mcp", Name: "test"}, Level: LevelOnce})
	assert.FileExists(t, path)
}

func TestFileStore_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")
	s, err := NewFileStore(path)
	require.NoError(t, err)

	list, err := s.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 0)
}

func TestFileStore_Delete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "grants.json")
	s, err := NewFileStore(path)
	require.NoError(t, err)

	ctx := context.Background()
	s.Save(ctx, Grant{Action: Action{Scope: "mcp", Name: "del"}, Level: LevelOnce})
	s.Delete(ctx, Action{Scope: "mcp", Name: "del"})

	list, _ := s.List(ctx)
	assert.Len(t, list, 0)
}

func TestMemoryStore_Close(t *testing.T) {
	s := NewMemoryStore()
	err := s.Close(context.Background())
	assert.NoError(t, err)
}

func TestFileStore_Close(t *testing.T) {
	s, err := NewFileStore(filepath.Join(t.TempDir(), "g.json"))
	require.NoError(t, err)
	err = s.Close(context.Background())
	assert.NoError(t, err)
}

func TestManagerCheck_AskFollowedByGrant(t *testing.T) {
	store := NewMemoryStore()
	m := NewManager("sess-1", store)

	var asked bool
	m.SetAskFunc(func(ctx context.Context, a Action) (Level, error) {
		asked = true
		return LevelSession, nil
	})

	m.SetRules([]Rule{
		{Scope: "mcp", Name: "*", Decision: DecisionAsk},
	})

	level, err := m.Check(context.Background(), Action{Scope: "mcp", Name: "tool"})
	require.NoError(t, err)
	assert.Equal(t, LevelSession, level)
	assert.True(t, asked)

	asked = false
	level, err = m.Check(context.Background(), Action{Scope: "mcp", Name: "tool"})
	require.NoError(t, err)
	assert.Equal(t, LevelSession, level)
	assert.False(t, asked, "should not ask again for granted action")
}

func TestMatchGlob(t *testing.T) {
	assert.True(t, matchGlob("mcp", "mcp"))
	assert.True(t, matchGlob("*", "anything"))
	assert.True(t, matchGlob("", "anything"))
	assert.True(t, matchGlob("mcp:*", "mcp:search"))
	assert.False(t, matchGlob("model", "mcp"))
}

func TestIntegration_AllowAllRule(t *testing.T) {
	store := NewMemoryStore()
	m := NewManager("sess-1", store)
	m.SetRules([]Rule{
		{Scope: "*", Name: "*", Decision: DecisionAllow, Level: LevelSession},
	})

	tools := []string{"search", "read", "write", "delete"}
	for _, name := range tools {
		level, err := m.Check(context.Background(), Action{Scope: "mcp", Name: name})
		require.NoError(t, err)
		assert.Equal(t, LevelSession, level)
	}
}
