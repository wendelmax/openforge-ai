package skill

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSkillDefaults(t *testing.T) {
	s := Skill{
		Name: "test",
	}
	assert.Equal(t, "test", s.Name)
	assert.Empty(t, s.Description)
}

func TestStepTypes(t *testing.T) {
	assert.Equal(t, StepType("prompt"), StepPrompt)
	assert.Equal(t, StepType("embed"), StepEmbed)
	assert.Equal(t, StepType("rerank"), StepRerank)
	assert.Equal(t, StepType("format"), StepFormat)
	assert.Equal(t, StepType("condition"), StepCond)
}

func TestSkillWithSteps(t *testing.T) {
	s := Skill{
		Name: "summarize",
		Steps: []Step{
			{Type: StepPrompt, Name: "step1", Model: "llama-3.2-3b", Output: "summary"},
			{Type: StepFormat, Name: "step2", Output: "result"},
		},
	}
	assert.Len(t, s.Steps, 2)
	assert.Equal(t, "llama-3.2-3b", s.Steps[0].Model)
	assert.Equal(t, "result", s.Steps[1].Output)
}

func TestStepConfig(t *testing.T) {
	s := Step{
		Type:   StepPrompt,
		Name:   "test",
		Output: "out",
		Config: map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  2048,
		},
	}
	assert.Equal(t, 0.7, s.Config["temperature"])
	assert.Equal(t, 2048, s.Config["max_tokens"])
}
