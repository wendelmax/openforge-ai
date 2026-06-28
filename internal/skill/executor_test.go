package skill

import (
	"context"
	"testing"

	"github.com/openforge-ai/openforge/runtime"
	"github.com/stretchr/testify/assert"
)

type mockRuntime struct{}

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
	return &runtime.InferenceResult{Text: "mock: " + prompt}, nil
}
func (m *mockRuntime) InferStream(ctx context.Context, modelID string, prompt string, params runtime.InferenceParams) (<-chan string, error) {
	return nil, nil
}
func (m *mockRuntime) Embed(ctx context.Context, modelID string, inputs []string, device string) (*runtime.EmbeddingResult, error) {
	n := len(inputs)
	embs := make([][]float32, n)
	for i := 0; i < n; i++ {
		embs[i] = []float32{0.1, 0.2, 0.3}
	}
	return &runtime.EmbeddingResult{Embeddings: embs}, nil
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
	return nil
}

func TestExecutor_Execute_PromptStep(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "test",
		Steps: []Step{
			{
				Type:   StepPrompt,
				Name:   "generate",
				Model:  "test-model",
				System: "You are a helpful assistant.",
				User:   "Write about {{.topic}}",
				Output: "result",
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, map[string]interface{}{
		"topic": "AI",
	})
	assert.NoError(t, err)
	assert.Contains(t, outputs, "result")
	result := outputs["result"].(string)
	assert.Contains(t, result, "mock:")
	assert.Contains(t, result, "AI")
}

func TestExecutor_Execute_EmbedStep(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "test-embed",
		Steps: []Step{
			{
				Type:   StepEmbed,
				Name:   "encode",
				Model:  "bge-small",
				Input:  "embed this text",
				Output: "vector",
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, nil)
	assert.NoError(t, err)
	assert.Contains(t, outputs, "vector")
	vec := outputs["vector"].([]float32)
	assert.Equal(t, []float32{0.1, 0.2, 0.3}, vec)
}

func TestExecutor_Execute_FormatStep(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "test-format",
		Steps: []Step{
			{
				Type:     StepFormat,
				Name:     "format",
				Template: "Hello, {{.name}}!",
				Output:   "greeting",
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, map[string]interface{}{
		"name": "World",
	})
	assert.NoError(t, err)
	assert.Contains(t, outputs, "greeting")
	assert.Equal(t, "Hello, World!", outputs["greeting"])
}

func TestExecutor_Execute_UnknownStep(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "test-unknown",
		Steps: []Step{
			{
				Type:   "unknown",
				Name:   "bad",
				Output: "out",
			},
		},
	}

	_, err := ex.Execute(context.Background(), skill, nil)
	assert.Error(t, err)
}

func TestExecutor_Execute_ContextCanceled(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "test-cancel",
		Steps: []Step{
			{
				Type:   StepPrompt,
				Name:   "generate",
				Model:  "test",
				Output: "result",
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := ex.Execute(ctx, skill, nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestExecutor_Execute_MultiStep(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "multi-step",
		Steps: []Step{
			{
				Type:     StepFormat,
				Name:     "greet",
				Template: "Hello {{.name}}",
				Output:   "greeting",
			},
			{
				Type:   StepPrompt,
				Name:   "respond",
				Model:  "test",
				System: "You say hello.",
				User:   "{{.greeting}}",
				Output: "response",
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, map[string]interface{}{
		"name": "World",
	})
	assert.NoError(t, err)
	assert.Contains(t, outputs, "greeting")
	assert.Contains(t, outputs, "response")
}

func TestExecutor_Execute_EmptyStep(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name:  "empty",
		Steps: []Step{},
	}

	outputs, err := ex.Execute(context.Background(), skill, map[string]interface{}{"x": 1})
	assert.NoError(t, err)
	assert.Equal(t, 1, outputs["inputs"].(map[string]interface{})["x"])
}

func TestRenderTemplate_MissingKey(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)
	result := ex.renderTemplate("hello {{.name}}", map[string]interface{}{})
	assert.Contains(t, result, "hello")
}

func TestRenderTemplate_Empty(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)
	result := ex.renderTemplate("", nil)
	assert.Equal(t, "", result)
}

func TestRenderTemplate_NoPlaceholders(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)
	result := ex.renderTemplate("static text", nil)
	assert.Equal(t, "static text", result)
}

func TestNewExecutor(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)
	assert.NotNil(t, ex)
}

func TestExecutePrompt_DefaultModel(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "default-model",
		Steps: []Step{
			{
				Type:   StepPrompt,
				Name:   "gen",
				Output: "result",
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, nil)
	assert.NoError(t, err)
	assert.Contains(t, outputs, "result")
}

func TestExecuteEmbed_EmptyInput(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "empty-embed",
		Steps: []Step{
			{
				Type:   StepEmbed,
				Name:   "encode",
				Output: "vector",
				Model:  "bge",
				Input:  "",
			},
		},
	}

	_, err := ex.Execute(context.Background(), skill, nil)
	assert.Error(t, err)
}

func TestExecuteRerankStep(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "test-rerank",
		Steps: []Step{
			{
				Type:   StepRerank,
				Name:   "rank",
				Model:  "bge-reranker",
				Query:  "test query",
				TopN:   2,
				Output: "results",
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, map[string]interface{}{
		"documents": []string{"doc1", "doc2", "doc3"},
	})
	assert.NoError(t, err)
	assert.Contains(t, outputs, "results")

	results, ok := outputs["results"].([]map[string]interface{})
	if !ok {
		t.Logf("results type: %T", outputs["results"])
	}
	_ = ok
	assert.NotNil(t, results)
}

func TestExecuteRerankStepMissingDocuments(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "test-rerank-missing-docs",
		Steps: []Step{
			{
				Type:   StepRerank,
				Name:   "rank",
				Model:  "bge-reranker",
				Query:  "test",
				Output: "results",
			},
		},
	}

	_, err := ex.Execute(context.Background(), skill, nil)
	assert.Error(t, err)
}

func TestExecuteRerankStepDefaultModel(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "test-rerank-default",
		Steps: []Step{
			{
				Type:   StepRerank,
				Name:   "rank",
				Query:  "test query",
				Output: "results",
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, map[string]interface{}{
		"documents": []string{"doc1", "doc2", "doc3"},
	})
	assert.NoError(t, err)
	assert.Contains(t, outputs, "results")
}

func TestExecuteConditionTrue(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "test-cond",
		Steps: []Step{
			{
				Type:     StepCond,
				Name:     "check",
				Template: "{{.inputs.value}}",
				Output:   "valid",
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, map[string]interface{}{
		"value": "true",
	})
	assert.NoError(t, err)
	assert.Equal(t, "true", outputs["valid"])
}

func TestExecuteConditionFalse(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "test-cond-false",
		Steps: []Step{
			{
				Type:     StepCond,
				Name:     "check",
				Template: "{{.inputs.value}}",
				Output:   "valid",
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, map[string]interface{}{
		"value": "false",
	})
	assert.NoError(t, err)
	assert.Equal(t, "false", outputs["valid"])
}

func TestExecuteConditionEquals(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "test-cond-eq",
		Steps: []Step{
			{
				Type:     StepCond,
				Name:     "check",
				Template: "{{.inputs.value}}",
				Output:   "valid",
				Config: map[string]interface{}{
					"equals": "yes",
				},
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, map[string]interface{}{
		"value": "yes",
	})
	assert.NoError(t, err)
	assert.Equal(t, "true", outputs["valid"])
}

func TestExecuteConditionNotEquals(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "test-cond-neq",
		Steps: []Step{
			{
				Type:     StepCond,
				Name:     "check",
				Template: "{{.inputs.value}}",
				Output:   "valid",
				Config: map[string]interface{}{
					"equals": "yes",
				},
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, map[string]interface{}{
		"value": "no",
	})
	assert.NoError(t, err)
	assert.Equal(t, "false", outputs["valid"])
}

func TestExecuteConditionEmpty(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "test-cond-empty",
		Steps: []Step{
			{
				Type:     StepCond,
				Name:     "check",
				Template: "{{.inputs.value}}",
				Output:   "valid",
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, map[string]interface{}{
		"value": "",
	})
	assert.NoError(t, err)
	assert.Equal(t, "false", outputs["valid"])
}

func TestCosineSimilarity(t *testing.T) {
	a := []float32{1, 0, 0}
	b := []float32{1, 0, 0}
	assert.InDelta(t, float32(1.0), cosineSimilarity(a, b), 0.01)

	c := []float32{0, 1, 0}
	assert.InDelta(t, float32(0.0), cosineSimilarity(a, c), 0.01)

	d := []float32{0.5, 0, 0}
	assert.InDelta(t, float32(1.0), cosineSimilarity(a, d), 0.01)
}

func TestCosineSimilarityEmpty(t *testing.T) {
	assert.Equal(t, float32(0), cosineSimilarity(nil, []float32{1, 2}))
	assert.Equal(t, float32(0), cosineSimilarity([]float32{1, 2}, []float32{1, 2, 3}))
}

func TestCosineSimilarityOrthogonal(t *testing.T) {
	a := []float32{1, 0, 0}
	b := []float32{0, 1, 0}
	assert.InDelta(t, float32(0), cosineSimilarity(a, b), 0.01)
}

func TestExecutePromptWithConfigTemp(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "temp-config",
		Steps: []Step{
			{
				Type:   StepPrompt,
				Name:   "gen",
				Model:  "test",
				User:   "hello",
				Output: "result",
				Config: map[string]interface{}{
					"temperature": 0.5,
				},
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, nil)
	assert.NoError(t, err)
	assert.Contains(t, outputs, "result")
}

func TestExecuteEmbed_DefaultModel(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "embed-default",
		Steps: []Step{
			{
				Type:  StepEmbed,
				Name:  "encode",
				Input: "some text",
				Output: "vector",
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, nil)
	assert.NoError(t, err)
	assert.Contains(t, outputs, "vector")
}

func TestExecuteFormat_InputFallback(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "format-fallback",
		Steps: []Step{
			{
				Type:  StepFormat,
				Name:  "fmt",
				Input: "{{.name}}",
				Output: "greeting",
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, map[string]interface{}{
		"name": "World",
	})
	assert.NoError(t, err)
	assert.Equal(t, "World", outputs["greeting"])
}

func TestExecuteRerank_NonStringDocuments(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "rerank-bad-docs",
		Steps: []Step{
			{
				Type:   StepRerank,
				Name:   "rank",
				Model:  "bge-reranker",
				Query:  "test",
				Output: "results",
			},
		},
	}

	_, err := ex.Execute(context.Background(), skill, map[string]interface{}{
		"documents": "not a slice",
	})
	assert.Error(t, err)
}

func TestRenderTemplate_InvalidSyntax(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)
	result := ex.renderTemplate("{{.name", nil)
	assert.Equal(t, "{{.name", result)
}

func TestRenderTemplate_ExecuteError(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)
	result := ex.renderTemplate("{{.Recursion.Recursion.Recursion}}", map[string]interface{}{
		"Recursion": map[string]interface{}{
			"Recursion": "value",
		},
	})
	assert.NotEmpty(t, result)
}

func TestExecuteCondition_ZeroInput(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "cond-zero",
		Steps: []Step{
			{
				Type:     StepCond,
				Name:     "check",
				Template: "{{.inputs.value}}",
				Output:   "valid",
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, map[string]interface{}{
		"value": "0",
	})
	assert.NoError(t, err)
	assert.Equal(t, "false", outputs["valid"])
}

func TestExecuteCondition_NoInput(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "cond-no",
		Steps: []Step{
			{
				Type:     StepCond,
				Name:     "check",
				Template: "{{.inputs.value}}",
				Output:   "valid",
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, map[string]interface{}{
		"value": "no",
	})
	assert.NoError(t, err)
	assert.Equal(t, "false", outputs["valid"])
}

func TestBuildStepInputs_StepInfo(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "step-info",
		Steps: []Step{
			{
				Type:   StepFormat,
				Name:   "fmt",
				Input:  "{{.step.name}}",
				Output: "out",
				Config: map[string]interface{}{"key": "val"},
			},
		},
	}

	outputs, err := ex.Execute(context.Background(), skill, nil)
	assert.NoError(t, err)
	assert.Equal(t, "fmt", outputs["out"])
}

func TestExecuteEmbed_NoEmbeddings(t *testing.T) {
	rt := &noEmbedMockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "no-emb",
		Steps: []Step{
			{
				Type:   StepEmbed,
				Name:   "encode",
				Model:  "test",
				Input:  "text",
				Output: "vec",
			},
		},
	}

	_, err := ex.Execute(context.Background(), skill, nil)
	assert.Error(t, err)
}

type noEmbedMockRuntime struct {
	mockRuntime
}

func (m *noEmbedMockRuntime) Embed(ctx context.Context, modelID string, inputs []string, device string) (*runtime.EmbeddingResult, error) {
	return &runtime.EmbeddingResult{Embeddings: nil}, nil
}

func TestExecutePrompt_InferError(t *testing.T) {
	rt := &inferErrorMock{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "infer-error",
		Steps: []Step{
			{
				Type:   StepPrompt,
				Name:   "gen",
				Model:  "test",
				User:   "hello",
				Output: "result",
			},
		},
	}

	_, err := ex.Execute(context.Background(), skill, nil)
	assert.Error(t, err)
}

type inferErrorMock struct {
	mockRuntime
}

func (m *inferErrorMock) Infer(ctx context.Context, modelID string, prompt string, params runtime.InferenceParams) (*runtime.InferenceResult, error) {
	return nil, assert.AnError
}

func TestExecuteRerank_EmptyQuery(t *testing.T) {
	rt := &mockRuntime{}
	ex := NewExecutor(rt)

	skill := Skill{
		Name: "rerank-empty-query",
		Steps: []Step{
			{
				Type:   StepRerank,
				Name:   "rank",
				Model:  "bge-reranker",
				Query:  "",
				Output: "results",
			},
		},
	}

	_, err := ex.Execute(context.Background(), skill, map[string]interface{}{
		"documents": []string{"doc1"},
	})
	assert.Error(t, err)
}
