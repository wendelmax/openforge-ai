// Package skill defines the skill execution pipeline.
package skill

import "context"

// StepType defines the type of a pipeline step.
type StepType string

const (
	// StepPrompt sends a prompt to an LLM and returns the response.
	StepPrompt StepType = "prompt"
	// StepEmbed generates embeddings for input text.
	StepEmbed StepType = "embed"
	// StepRerank reranks documents using a cross-encoder (not yet implemented).
	StepRerank StepType = "rerank"
	// StepFormat formats output using a Go template.
	StepFormat StepType = "format"
	// StepCond conditionally executes steps (not yet implemented).
	StepCond StepType = "condition"
)

// Step is a single pipeline step within a skill.
type Step struct {
	Type     StepType `yaml:"type"`
	Name     string   `yaml:"name"`
	ID       string   `yaml:"id,omitempty"`
	Model    string   `yaml:"model,omitempty"`
	Input    string   `yaml:"input,omitempty"`
	Output   string   `yaml:"output"`
	System   string   `yaml:"system,omitempty"`
	User     string   `yaml:"user,omitempty"`
	Template string   `yaml:"template,omitempty"`
	Query    string   `yaml:"query,omitempty"`
	TopN     int      `yaml:"top_n,omitempty"`
	Config   map[string]interface{} `yaml:"config,omitempty"`
}

// Skill defines a named pipeline of steps to execute.
type Skill struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version,omitempty"`
	Steps       []Step `yaml:"steps"`
}

//go:generate mockery --name ExecutorInterface --output ../mocks --outpkg mocks

// ExecutorInterface defines the contract for executing skills.
type ExecutorInterface interface {
	Execute(ctx context.Context, skill Skill, inputs map[string]interface{}) (map[string]interface{}, error)
}
