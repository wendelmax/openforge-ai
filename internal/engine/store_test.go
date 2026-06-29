package engine

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStore_MemoryCreate(t *testing.T) {
	s := NewMemorySessionStore()
	ctx := context.Background()
	session := &Session{ModelID: "test"}
	err := s.Create(ctx, session)
	assert.NoError(t, err)
	assert.NotEmpty(t, session.ID)
}

func TestStore_MemoryGet(t *testing.T) {
	s := NewMemorySessionStore()
	ctx := context.Background()
	session := &Session{ModelID: "test"}
	s.Create(ctx, session)
	got, err := s.Get(ctx, session.ID)
	assert.NoError(t, err)
	assert.Equal(t, session.ID, got.ID)
}

func TestStore_MemoryGetNotFound(t *testing.T) {
	s := NewMemorySessionStore()
	_, err := s.Get(context.Background(), "nope")
	assert.Error(t, err)
}

func TestStore_MemoryUpdateDeleteList(t *testing.T) {
	s := NewMemorySessionStore()
	ctx := context.Background()
	s1 := &Session{ModelID: "a"}
	s2 := &Session{ModelID: "b"}
	s.Create(ctx, s1)
	s.Create(ctx, s2)
	list, _ := s.List(ctx)
	assert.Len(t, list, 2)
	s1.ModelID = "updated"
	s.Update(ctx, s1)
	got, _ := s.Get(ctx, s1.ID)
	assert.Equal(t, "updated", got.ModelID)
	s.Delete(ctx, s2.ID)
	list2, _ := s.List(ctx)
	assert.Len(t, list2, 1)
}

func TestStore_FileCreate(t *testing.T) {
	dir := t.TempDir()
	s, err := NewFileSessionStore(dir)
	assert.NoError(t, err)
	session := &Session{ModelID: "test"}
	s.Create(context.Background(), session)
	assert.NotEmpty(t, session.ID)
}

func TestStore_FilePersistence(t *testing.T) {
	dir := t.TempDir()
	s1, _ := NewFileSessionStore(dir)
	ctx := context.Background()
	session := &Session{ModelID: "test"}
	s1.Create(ctx, session)
	s1.Close(ctx)
	s2, _ := NewFileSessionStore(dir)
	got, err := s2.Get(ctx, session.ID)
	assert.NoError(t, err)
	assert.Equal(t, session.ID, got.ID)
}

func TestStore_FileNestedDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "a", "b")
	s, err := NewFileSessionStore(dir)
	assert.NoError(t, err)
	assert.NotNil(t, s)
}

func TestStore_InvalidDir(t *testing.T) {
	t.Skip("/proc not available in test environment")
}

func TestStore_NewWithStore(t *testing.T) {
	rt := &mockRuntime{}
	store := NewMemorySessionStore()
	e := NewWithStore(rt, store)
	assert.NotNil(t, e)
	assert.Equal(t, store, e.Store())
}

func TestStore_NewWithFileStore(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewFileSessionStore(dir)
	rt := &mockRuntime{}
	e := NewWithStore(rt, store)
	ctx := context.Background()
	s, _ := e.CreateSession(ctx, "test")
	got, _ := e.GetSession(ctx, s.ID)
	assert.Equal(t, s.ID, got.ID)
}

func TestStore_NewWithSQLiteStore(t *testing.T) {
	rt := &mockRuntime{}
	store, err := NewSQLiteSessionStore(filepath.Join(t.TempDir(), "s.db"))
	if err != nil { t.Skip("sqlite not available") }
	defer store.Close(context.Background())
	e := NewWithStore(rt, store)
	assert.NotNil(t, e)
}

var _ = time.Now
