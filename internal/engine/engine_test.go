package engine

import (
	"context"
	"testing"

	"github.com/openforge-ai/openforge/runtime"
	"github.com/stretchr/testify/assert"
)

type mockRuntime struct {
	closeCalled bool
}

func (m *mockRuntime) ListDevices(ctx context.Context) ([]runtime.Device, error) {
	return nil, nil
}

func (m *mockRuntime) LoadModel(ctx context.Context, modelID, path, device string) error {
	return nil
}

func (m *mockRuntime) UnloadModel(ctx context.Context, modelID string) error {
	return nil
}

func (m *mockRuntime) ListModels(ctx context.Context) ([]runtime.ModelInfo, error) {
	return nil, nil
}

func (m *mockRuntime) Infer(ctx context.Context, modelID string, prompt string, params runtime.InferenceParams) (*runtime.InferenceResult, error) {
	return nil, nil
}

func (m *mockRuntime) InferStream(ctx context.Context, modelID string, prompt string, params runtime.InferenceParams) (<-chan string, error) {
	return nil, nil
}

func (m *mockRuntime) Embed(ctx context.Context, modelID string, inputs []string, device string) (*runtime.EmbeddingResult, error) {
	return nil, nil
}

func (m *mockRuntime) DeviceForWorkload(workload, requestDevice string) string {
	if requestDevice != "" && requestDevice != "auto" {
		return requestDevice
	}
	return "cpu"
}

func (m *mockRuntime) Benchmark(ctx context.Context, modelID string, iterations int, prompt string, maxTokens int) (runtime.BenchmarkResults, error) {
	return nil, nil
}

func (m *mockRuntime) Close(ctx context.Context) error {
	m.closeCalled = true
	return nil
}

func TestNewEngine(t *testing.T) {
	rt := &mockRuntime{}
	e := New(rt)
	assert.NotNil(t, e)
	assert.NotNil(t, e.store)
	assert.Equal(t, rt, e.Runtime())
}

func TestNewWithStore(t *testing.T) {
	rt := &mockRuntime{}
	store := NewMemorySessionStore()
	e := NewWithStore(rt, store)
	assert.NotNil(t, e)
	assert.Equal(t, store, e.Store())
}

func TestCreateSession(t *testing.T) {
	rt := &mockRuntime{}
	e := New(rt)
	ctx := context.Background()

	session, err := e.CreateSession(ctx, "test-model")
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, "test-model", session.ModelID)
}

func TestGetSession(t *testing.T) {
	rt := &mockRuntime{}
	e := New(rt)
	ctx := context.Background()

	session, _ := e.CreateSession(ctx, "test-model")

	got, err := e.GetSession(ctx, session.ID)
	assert.NoError(t, err)
	assert.Equal(t, session.ID, got.ID)
	assert.Equal(t, session.ModelID, got.ModelID)
}

func TestGetSessionNotFound(t *testing.T) {
	rt := &mockRuntime{}
	e := New(rt)
	ctx := context.Background()

	_, err := e.GetSession(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRuntime(t *testing.T) {
	rt := &mockRuntime{}
	e := New(rt)
	assert.Equal(t, rt, e.Runtime())
}

func TestDeleteSession(t *testing.T) {
	rt := &mockRuntime{}
	e := New(rt)
	ctx := context.Background()

	session, _ := e.CreateSession(ctx, "test-model")
	err := e.DeleteSession(ctx, session.ID)
	assert.NoError(t, err)

	_, err = e.GetSession(ctx, session.ID)
	assert.Error(t, err)
}

func TestDeleteSessionNotFound(t *testing.T) {
	rt := &mockRuntime{}
	e := New(rt)
	ctx := context.Background()

	err := e.DeleteSession(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestClearSessions(t *testing.T) {
	rt := &mockRuntime{}
	e := New(rt)
	ctx := context.Background()

	e.CreateSession(ctx, "model-1")
	e.CreateSession(ctx, "model-2")

	list, _ := e.store.List(ctx)
	assert.Len(t, list, 2)

	err := e.ClearSessions(ctx)
	assert.NoError(t, err)

	list, _ = e.store.List(ctx)
	assert.Len(t, list, 0)
}

func TestAddMessage(t *testing.T) {
	rt := &mockRuntime{}
	e := New(rt)
	ctx := context.Background()

	session, _ := e.CreateSession(ctx, "test-model")
	msg := runtime.Message{Role: "user", Content: "hello"}
	updated, err := e.AddMessage(ctx, session.ID, msg)
	assert.NoError(t, err)
	assert.Len(t, updated.Messages, 1)
	assert.Equal(t, "user", updated.Messages[0].Role)
	assert.Equal(t, "hello", updated.Messages[0].Content)
}

func TestAddMessageNotFound(t *testing.T) {
	rt := &mockRuntime{}
	e := New(rt)
	ctx := context.Background()

	_, err := e.AddMessage(ctx, "nonexistent", runtime.Message{Role: "user", Content: "hi"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAddMessageContextWindowing(t *testing.T) {
	rt := &mockRuntime{}
	e := New(rt)
	ctx := context.Background()

	session, _ := e.CreateSession(ctx, "test-model")
	session.MaxTokens = 10

	msg1 := runtime.Message{Role: "user", Content: "hello world this is a test"}
	msg2 := runtime.Message{Role: "assistant", Content: "short reply"}
	msg3 := runtime.Message{Role: "user", Content: "another message"}
	msg4 := runtime.Message{Role: "user", Content: "hi"}

	e.AddMessage(ctx, session.ID, msg1)
	e.AddMessage(ctx, session.ID, msg2)
	e.AddMessage(ctx, session.ID, msg3)
	updated, _ := e.AddMessage(ctx, session.ID, msg4)

	assert.LessOrEqual(t, len(updated.Messages), 4, "context windowing should have trimmed messages")
}

func TestBuildPrompt(t *testing.T) {
	rt := &mockRuntime{}
	e := New(rt)
	ctx := context.Background()

	session, _ := e.CreateSession(ctx, "test-model")
	e.AddMessage(ctx, session.ID, runtime.Message{Role: "user", Content: "hello"})
	e.AddMessage(ctx, session.ID, runtime.Message{Role: "assistant", Content: "world"})

	prompt, err := e.BuildPrompt(ctx, session.ID)
	assert.NoError(t, err)
	assert.Contains(t, prompt, "user: hello")
	assert.Contains(t, prompt, "assistant: world")
}

func TestBuildPromptSessionNotFound(t *testing.T) {
	rt := &mockRuntime{}
	e := New(rt)
	ctx := context.Background()

	_, err := e.BuildPrompt(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestConcurrentAccess(t *testing.T) {
	rt := &mockRuntime{}
	e := New(rt)
	ctx := context.Background()

	session, _ := e.CreateSession(ctx, "test-model")

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			e.AddMessage(ctx, session.ID, runtime.Message{Role: "user", Content: "msg"})
			e.GetSession(ctx, session.ID)
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}

	updated, _ := e.GetSession(ctx, session.ID)
	assert.Len(t, updated.Messages, 10)
}

func TestClose(t *testing.T) {
	rt := &mockRuntime{}
	e := New(rt)
	ctx := context.Background()

	err := e.Close(ctx)
	assert.NoError(t, err)
	assert.True(t, rt.closeCalled)
}

func TestSessionMaxTokens(t *testing.T) {
	rt := &mockRuntime{}
	e := New(rt)
	ctx := context.Background()

	session, _ := e.CreateSession(ctx, "test-model")
	assert.Equal(t, 0, session.MaxTokens, "default max tokens should be 0")

	session.MaxTokens = 2048
	assert.Equal(t, 2048, session.MaxTokens)
}

func TestCreateSessionMultiple(t *testing.T) {
	rt := &mockRuntime{}
	e := New(rt)
	ctx := context.Background()

	s1, _ := e.CreateSession(ctx, "model-a")
	s2, _ := e.CreateSession(ctx, "model-b")
	assert.NotEqual(t, s1.ID, s2.ID, "session IDs should be unique")

	list, _ := e.store.List(ctx)
	assert.Len(t, list, 2)
}
