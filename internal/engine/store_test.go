package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/openforge-ai/openforge/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemorySessionStore_Create(t *testing.T) {
	s := NewMemorySessionStore()
	ctx := context.Background()

	session := &Session{
		ModelID:  "test-model",
		Messages: []runtime.Message{{Role: "user", Content: "hello"}},
	}
	err := s.Create(ctx, session)
	require.NoError(t, err)
	assert.NotEmpty(t, session.ID)
}

func TestMemorySessionStore_CreateWithID(t *testing.T) {
	s := NewMemorySessionStore()
	ctx := context.Background()

	session := &Session{
		ID:      "my-id",
		ModelID: "test-model",
	}
	err := s.Create(ctx, session)
	require.NoError(t, err)
	assert.Equal(t, "my-id", session.ID)
}

func TestMemorySessionStore_GetFound(t *testing.T) {
	s := NewMemorySessionStore()
	ctx := context.Background()

	session := &Session{ModelID: "test-model"}
	s.Create(ctx, session)

	got, err := s.Get(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, got.ID)
}

func TestMemorySessionStore_GetNotFound(t *testing.T) {
	s := NewMemorySessionStore()
	ctx := context.Background()

	_, err := s.Get(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMemorySessionStore_Update(t *testing.T) {
	s := NewMemorySessionStore()
	ctx := context.Background()

	session := &Session{ModelID: "test-model"}
	s.Create(ctx, session)

	session.ModelID = "updated-model"
	err := s.Update(ctx, session)
	require.NoError(t, err)

	got, _ := s.Get(ctx, session.ID)
	assert.Equal(t, "updated-model", got.ModelID)
}

func TestMemorySessionStore_UpdateNotFound(t *testing.T) {
	s := NewMemorySessionStore()
	ctx := context.Background()

	err := s.Update(ctx, &Session{ID: "nonexistent"})
	assert.Error(t, err)
}

func TestMemorySessionStore_Delete(t *testing.T) {
	s := NewMemorySessionStore()
	ctx := context.Background()

	session := &Session{ModelID: "test-model"}
	s.Create(ctx, session)

	err := s.Delete(ctx, session.ID)
	require.NoError(t, err)

	_, err = s.Get(ctx, session.ID)
	assert.Error(t, err)
}

func TestMemorySessionStore_DeleteNotFound(t *testing.T) {
	s := NewMemorySessionStore()
	ctx := context.Background()

	err := s.Delete(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestMemorySessionStore_List(t *testing.T) {
	s := NewMemorySessionStore()
	ctx := context.Background()

	s.Create(ctx, &Session{ModelID: "a", CreatedAt: time.Now()})
	s.Create(ctx, &Session{ModelID: "b", CreatedAt: time.Now().Add(time.Second)})

	list, err := s.List(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestMemorySessionStore_ListEmpty(t *testing.T) {
	s := NewMemorySessionStore()
	ctx := context.Background()

	list, err := s.List(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 0)
}

func TestMemorySessionStore_Close(t *testing.T) {
	s := NewMemorySessionStore()
	ctx := context.Background()

	s.Create(ctx, &Session{ModelID: "a"})
	err := s.Close(ctx)
	require.NoError(t, err)

	list, _ := s.List(ctx)
	assert.Len(t, list, 0)
}

func TestFileSessionStore_CreateAndGet(t *testing.T) {
	dir := t.TempDir()
	s, err := NewFileSessionStore(dir)
	require.NoError(t, err)
	ctx := context.Background()

	session := &Session{ModelID: "test-model"}
	err = s.Create(ctx, session)
	require.NoError(t, err)

	got, err := s.Get(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, got.ID)
	assert.Equal(t, "test-model", got.ModelID)
}

func TestFileSessionStore_GetNotFound(t *testing.T) {
	dir := t.TempDir()
	s, err := NewFileSessionStore(dir)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = s.Get(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFileSessionStore_Delete(t *testing.T) {
	dir := t.TempDir()
	s, err := NewFileSessionStore(dir)
	require.NoError(t, err)
	ctx := context.Background()

	session := &Session{ModelID: "test-model"}
	s.Create(ctx, session)

	err = s.Delete(ctx, session.ID)
	require.NoError(t, err)

	_, err = s.Get(ctx, session.ID)
	assert.Error(t, err)
}

func TestFileSessionStore_Update(t *testing.T) {
	dir := t.TempDir()
	s, err := NewFileSessionStore(dir)
	require.NoError(t, err)
	ctx := context.Background()

	session := &Session{ModelID: "test-model"}
	s.Create(ctx, session)

	session.ModelID = "updated"
	err = s.Update(ctx, session)
	require.NoError(t, err)

	got, _ := s.Get(ctx, session.ID)
	assert.Equal(t, "updated", got.ModelID)
}

func TestFileSessionStore_List(t *testing.T) {
	dir := t.TempDir()
	s, err := NewFileSessionStore(dir)
	require.NoError(t, err)
	ctx := context.Background()

	s.Create(ctx, &Session{ModelID: "a"})
	s.Create(ctx, &Session{ModelID: "b"})

	list, err := s.List(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestFileSessionStore_ListEmpty(t *testing.T) {
	dir := t.TempDir()
	s, err := NewFileSessionStore(dir)
	require.NoError(t, err)
	ctx := context.Background()

	list, err := s.List(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 0)
}

func TestFileSessionStore_Close(t *testing.T) {
	dir := t.TempDir()
	s, err := NewFileSessionStore(dir)
	require.NoError(t, err)
	ctx := context.Background()

	s.Create(ctx, &Session{ModelID: "a"})
	err = s.Close(ctx)
	require.NoError(t, err)
}

func TestFileSessionStore_PersistenceAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	s1, err := NewFileSessionStore(dir)
	require.NoError(t, err)
	session := &Session{ModelID: "persist-test", Messages: []runtime.Message{{Role: "user", Content: "hello"}}}
	s1.Create(ctx, session)
	s1.Close(ctx)

	s2, err := NewFileSessionStore(dir)
	require.NoError(t, err)
	got, err := s2.Get(ctx, session.ID)
	require.NoError(t, err)
	assert.Equal(t, "persist-test", got.ModelID)
	assert.Len(t, got.Messages, 1)
}

func TestFileSessionStore_DirCreation(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "sessions")
	s, err := NewFileSessionStore(dir)
	require.NoError(t, err)
	assert.DirExists(t, dir)

	ctx := context.Background()
	session := &Session{ModelID: "test"}
	err = s.Create(ctx, session)
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(dir, session.ID+".json"))
}

func TestEngineWithFileStore(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileSessionStore(dir)
	require.NoError(t, err)

	rt := &mockRuntime{}
	e := NewWithStore(rt, store)
	ctx := context.Background()

	session, err := e.CreateSession(ctx, "file-model")
	require.NoError(t, err)

	msg := runtime.Message{Role: "user", Content: "persisted?"}
	updated, err := e.AddMessage(ctx, session.ID, msg)
	require.NoError(t, err)
	assert.Len(t, updated.Messages, 1)

	// Verify file was written
	assert.FileExists(t, filepath.Join(dir, session.ID+".json"))
}

func TestFileSessionStore_DeleteNotFound(t *testing.T) {
	dir := t.TempDir()
	s, err := NewFileSessionStore(dir)
	require.NoError(t, err)
	ctx := context.Background()

	err = s.Delete(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFileSessionStore_UpdateNotFound(t *testing.T) {
	dir := t.TempDir()
	s, err := NewFileSessionStore(dir)
	require.NoError(t, err)
	ctx := context.Background()

	err = s.Update(ctx, &Session{ID: "nonexistent"})
	// Update creates or overwrites the file
	got, err2 := s.Get(ctx, "nonexistent")
	// After update, the file should exist
	assert.NoError(t, err2)
	assert.Equal(t, "nonexistent", got.ID)
}

func TestMemoryStore_ListOrder(t *testing.T) {
	s := NewMemorySessionStore()
	ctx := context.Background()

	now := time.Now()
	s.Create(ctx, &Session{ModelID: "first", CreatedAt: now})
	s.Create(ctx, &Session{ModelID: "second", CreatedAt: now.Add(2 * time.Second)})
	s.Create(ctx, &Session{ModelID: "third", CreatedAt: now.Add(time.Second)})

	list, _ := s.List(ctx)
	assert.Len(t, list, 3)
	assert.Equal(t, "first", list[0].ModelID)
	assert.Equal(t, "third", list[1].ModelID)
	assert.Equal(t, "second", list[2].ModelID)
}

func TestFileStoreManagerIntegration(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileSessionStore(dir)
	require.NoError(t, err)

	ctx := context.Background()
	rt := &mockRuntime{}
	e := NewWithStore(rt, store)

	s1, _ := e.CreateSession(ctx, "m1")
	e.AddMessage(ctx, s1.ID, runtime.Message{Role: "user", Content: "hi"})

	s2, _ := e.CreateSession(ctx, "m2")
	e.AddMessage(ctx, s2.ID, runtime.Message{Role: "user", Content: "bye"})

	// Verify both sessions persisted
	file1 := filepath.Join(dir, s1.ID+".json")
	file2 := filepath.Join(dir, s2.ID+".json")
	assert.FileExists(t, file1)
	assert.FileExists(t, file2)

	// Reopen with new store and verify data survives
	store2, err := NewFileSessionStore(dir)
	require.NoError(t, err)
	e2 := NewWithStore(rt, store2)

	g1, err := e2.GetSession(ctx, s1.ID)
	require.NoError(t, err)
	assert.Len(t, g1.Messages, 1)
	assert.Equal(t, "hi", g1.Messages[0].Content)

	g2, err := e2.GetSession(ctx, s2.ID)
	require.NoError(t, err)
	assert.Len(t, g2.Messages, 1)
	assert.Equal(t, "bye", g2.Messages[0].Content)
}

func TestNewFileSessionStore_InvalidDir(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("skipping permission test as root")
	}
	_, err := NewFileSessionStore("/proc/readonly/sessions")
	assert.Error(t, err)
}
