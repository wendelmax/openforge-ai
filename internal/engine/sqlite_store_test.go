package engine

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/openforge-ai/openforge/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteSessionStore_CreateAndGet(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSQLiteSessionStore(filepath.Join(dir, "sessions.db"))
	require.NoError(t, err)
	ctx := context.Background()

	session := &Session{
		ID:      "test-id",
		ModelID: "test-model",
		Messages: []runtime.Message{
			{Role: "user", Content: "hello"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = s.Create(ctx, session)
	require.NoError(t, err)

	got, err := s.Get(ctx, "test-id")
	require.NoError(t, err)
	assert.Equal(t, "test-id", got.ID)
	assert.Equal(t, "test-model", got.ModelID)
	assert.Len(t, got.Messages, 1)
	assert.Equal(t, "hello", got.Messages[0].Content)
}

func TestSQLiteSessionStore_GetNotFound(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSQLiteSessionStore(filepath.Join(dir, "sessions.db"))
	require.NoError(t, err)
	ctx := context.Background()

	_, err = s.Get(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSQLiteSessionStore_Update(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSQLiteSessionStore(filepath.Join(dir, "sessions.db"))
	require.NoError(t, err)
	ctx := context.Background()

	session := &Session{
		ID:      "update-id",
		ModelID: "original",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.Create(ctx, session)

	session.ModelID = "updated"
	session.Messages = []runtime.Message{{Role: "user", Content: "updated msg"}}
	session.UpdatedAt = time.Now()
	err = s.Update(ctx, session)
	require.NoError(t, err)

	got, _ := s.Get(ctx, "update-id")
	assert.Equal(t, "updated", got.ModelID)
	assert.Len(t, got.Messages, 1)
	assert.Equal(t, "updated msg", got.Messages[0].Content)
}

func TestSQLiteSessionStore_Delete(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSQLiteSessionStore(filepath.Join(dir, "sessions.db"))
	require.NoError(t, err)
	ctx := context.Background()

	s.Create(ctx, &Session{ID: "del-id", ModelID: "m", CreatedAt: time.Now(), UpdatedAt: time.Now()})

	err = s.Delete(ctx, "del-id")
	require.NoError(t, err)

	_, err = s.Get(ctx, "del-id")
	assert.Error(t, err)
}

func TestSQLiteSessionStore_List(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSQLiteSessionStore(filepath.Join(dir, "sessions.db"))
	require.NoError(t, err)
	ctx := context.Background()

	now := time.Now()
	s.Create(ctx, &Session{ID: "a", ModelID: "m1", CreatedAt: now, UpdatedAt: now})
	s.Create(ctx, &Session{ID: "b", ModelID: "m2", CreatedAt: now.Add(time.Second), UpdatedAt: now})

	list, err := s.List(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "a", list[0].ID)
	assert.Equal(t, "b", list[1].ID)
}

func TestSQLiteSessionStore_CreateEmptyID(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSQLiteSessionStore(filepath.Join(dir, "sessions.db"))
	require.NoError(t, err)
	ctx := context.Background()

	err = s.Create(ctx, &Session{ModelID: "m", CreatedAt: time.Now(), UpdatedAt: time.Now()})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID is required")
}

func TestSQLiteSessionStore_PersistenceAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "sessions.db")
	ctx := context.Background()

	s1, err := NewSQLiteSessionStore(dbPath)
	require.NoError(t, err)
	s1.Create(ctx, &Session{
		ID:      "persist-id",
		ModelID: "m",
		Messages: []runtime.Message{{Role: "user", Content: "hello"}},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	s1.Close(ctx)

	s2, err := NewSQLiteSessionStore(dbPath)
	require.NoError(t, err)
	got, err := s2.Get(ctx, "persist-id")
	require.NoError(t, err)
	assert.Equal(t, "m", got.ModelID)
	assert.Len(t, got.Messages, 1)
	assert.Equal(t, "hello", got.Messages[0].Content)
}

func TestSQLiteSessionStore_UpdateNotFound(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSQLiteSessionStore(filepath.Join(dir, "sessions.db"))
	require.NoError(t, err)
	ctx := context.Background()

	err = s.Update(ctx, &Session{ID: "nonexistent"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSQLiteSessionStore_DeleteNotFound(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSQLiteSessionStore(filepath.Join(dir, "sessions.db"))
	require.NoError(t, err)
	ctx := context.Background()

	err = s.Delete(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSQliteSessionStore_ListEmpty(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSQLiteSessionStore(filepath.Join(dir, "sessions.db"))
	require.NoError(t, err)
	ctx := context.Background()

	list, err := s.List(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 0)
}

func TestEngineWithSQLiteStore(t *testing.T) {
	dir := t.TempDir()
	store, err := NewSQLiteSessionStore(filepath.Join(dir, "sessions.db"))
	require.NoError(t, err)

	rt := &mockRuntime{}
	e := NewWithStore(rt, store)
	ctx := context.Background()

	session, err := e.CreateSession(ctx, "sqlite-model")
	require.NoError(t, err)

	msg := runtime.Message{Role: "user", Content: "sqlite test"}
	updated, err := e.AddMessage(ctx, session.ID, msg)
	require.NoError(t, err)
	assert.Len(t, updated.Messages, 1)

	got, err := e.GetSession(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, "sqlite-model", got.ModelID)
}

func TestSQLiteSessionStore_Close(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSQLiteSessionStore(filepath.Join(dir, "sessions.db"))
	require.NoError(t, err)
	ctx := context.Background()

	s.Create(ctx, &Session{ID: "close-id", ModelID: "m", CreatedAt: time.Now(), UpdatedAt: time.Now()})
	err = s.Close(ctx)
	require.NoError(t, err)

	// After close, operations should fail
	_, err = s.Get(ctx, "close-id")
	assert.Error(t, err)
}
