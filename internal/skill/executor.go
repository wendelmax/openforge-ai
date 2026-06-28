package skill

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"text/template"

	"github.com/openforge-ai/openforge/runtime"
)

// Executor runs skill pipelines by executing each step in order.
type Executor struct {
	runtime runtime.Runtime
}

// NewExecutor creates a new Executor backed by the given runtime.
func NewExecutor(rt runtime.Runtime) *Executor {
	return &Executor{runtime: rt}
}

// Execute runs all steps in the skill pipeline, passing outputs between steps.
func (e *Executor) Execute(ctx context.Context, skill Skill, inputs map[string]interface{}) (map[string]interface{}, error) {
	outputs := make(map[string]interface{})
	outputs["inputs"] = inputs

	for i, step := range skill.Steps {
		select {
		case <-ctx.Done():
			return outputs, ctx.Err()
		default:
		}

		stepInputs := e.buildStepInputs(step, outputs)
		result, err := e.executeStep(ctx, step, stepInputs)
		if err != nil {
			return outputs, fmt.Errorf("step %d (%s): %w", i+1, step.Name, err)
		}

		if step.Output != "" {
			outputs[step.Output] = result
		}
	}

	return outputs, nil
}

func (e *Executor) executeStep(ctx context.Context, step Step, inputs map[string]interface{}) (interface{}, error) {
	switch step.Type {
	case StepPrompt:
		return e.executePrompt(ctx, step, inputs)
	case StepEmbed:
		return e.executeEmbed(ctx, step, inputs)
	case StepRerank:
		return e.executeRerank(ctx, step, inputs)
	case StepFormat:
		return e.executeFormat(step, inputs)
	case StepCond:
		return e.executeCondition(step, inputs)
	default:
		return nil, fmt.Errorf("unknown step type: %s", step.Type)
	}
}

func (e *Executor) executePrompt(ctx context.Context, step Step, inputs map[string]interface{}) (string, error) {
	model := step.Model
	if model == "" {
		model = "llama-3.2-3b"
	}

	system := e.renderTemplate(step.System, inputs)
	user := e.renderTemplate(step.User, inputs)

	prompt := ""
	if system != "" {
		prompt += "system: " + system + "\n"
	}
	if user != "" {
		prompt += "user: " + user + "\n"
	}

	params := runtime.InferenceParams{
		MaxTokens: 2048,
		Temperature: 0.7,
	}

	if temp, ok := step.Config["temperature"].(float64); ok {
		params.Temperature = float32(temp)
	}

	result, err := e.runtime.Infer(ctx, model, prompt, params)
	if err != nil {
		return "", fmt.Errorf("inference: %w", err)
	}

	return result.Text, nil
}

func (e *Executor) executeEmbed(ctx context.Context, step Step, inputs map[string]interface{}) ([]float32, error) {
	model := step.Model
	if model == "" {
		model = "bge-small-en-v1.5"
	}

	input := e.renderTemplate(step.Input, inputs)
	if input == "" {
		return nil, fmt.Errorf("embed input is empty")
	}

	result, err := e.runtime.Embed(ctx, model, []string{input}, "")
	if err != nil {
		return nil, fmt.Errorf("embedding: %w", err)
	}

	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return result.Embeddings[0], nil
}

func (e *Executor) executeFormat(step Step, inputs map[string]interface{}) (string, error) {
	tmpl := step.Template
	if tmpl == "" {
		tmpl = step.Input
	}
	return e.renderTemplate(tmpl, inputs), nil
}

func (e *Executor) executeRerank(ctx context.Context, step Step, inputs map[string]interface{}) (interface{}, error) {
	model := step.Model
	if model == "" {
		model = "bge-reranker-v2-m3"
	}

	query := e.renderTemplate(step.Query, inputs)
	if query == "" {
		return nil, fmt.Errorf("rerank query is empty")
	}

	documentsRaw, ok := inputs["documents"]
	if !ok {
		return nil, fmt.Errorf("rerank requires 'documents' input")
	}

	docs, ok := documentsRaw.([]string)
	if !ok {
		docsVal := fmt.Sprintf("%v", documentsRaw)
		_ = docsVal
		return nil, fmt.Errorf("rerank documents must be a string slice")
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("rerank documents list is empty")
	}

	allTexts := append([]string{query}, docs...)
	embResult, err := e.runtime.Embed(ctx, model, allTexts, "")
	if err != nil {
		return nil, fmt.Errorf("rerank embed: %w", err)
	}

	if len(embResult.Embeddings) != len(allTexts) {
		return nil, fmt.Errorf("rerank: expected %d embeddings, got %d", len(allTexts), len(embResult.Embeddings))
	}

	queryEmb := embResult.Embeddings[0]
	type scoredDoc struct {
		Index int     `json:"index"`
		Score float32 `json:"score"`
		Text  string  `json:"text"`
	}
	scored := make([]scoredDoc, len(docs))
	for i, docEmb := range embResult.Embeddings[1:] {
		scored[i] = scoredDoc{
			Index: i,
			Score: cosineSimilarity(queryEmb, docEmb),
			Text:  docs[i],
		}
	}

	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].Score > scored[i].Score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	topN := step.TopN
	if topN <= 0 || topN > len(scored) {
		topN = len(scored)
	}

	result := scored[:topN]
	generic := make([]map[string]interface{}, len(result))
	for i, s := range result {
		generic[i] = map[string]interface{}{
			"index": s.Index,
			"score": s.Score,
			"text":  s.Text,
		}
	}
	return generic, nil
}

func (e *Executor) executeCondition(step Step, inputs map[string]interface{}) (string, error) {
	tmpl := step.Template
	if tmpl == "" {
		tmpl = step.Input
	}
	result := e.renderTemplate(tmpl, inputs)

	if step.Config != nil {
		if expected, ok := step.Config["equals"]; ok {
			if result == fmt.Sprintf("%v", expected) {
				return "true", nil
			}
			return "false", nil
		}
	}

	if result == "" || result == "false" || result == "0" || result == "no" {
		return "false", nil
	}
	return "true", nil
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return float32(dot / denom)
}

func (e *Executor) renderTemplate(tmpl string, data interface{}) string {
	if tmpl == "" {
		return ""
	}
	t, err := template.New("").Option("missingkey=zero").Parse(tmpl)
	if err != nil {
		return tmpl
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return tmpl
	}
	return buf.String()
}

func (e *Executor) buildStepInputs(step Step, outputs map[string]interface{}) map[string]interface{} {
	inputs := make(map[string]interface{})

	for k, v := range outputs {
		inputs[k] = v
	}

	if m, ok := outputs["inputs"].(map[string]interface{}); ok {
		for k, v := range m {
			if _, exists := inputs[k]; !exists {
				inputs[k] = v
			}
		}
	}

	inputs["step"] = map[string]interface{}{
		"name":   step.Name,
		"type":   string(step.Type),
		"config": step.Config,
	}

	return inputs
}
