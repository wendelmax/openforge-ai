//go:build integration

package agent

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/openforge-ai/openforge/internal/pm"
	"github.com/openforge-ai/openforge/internal/pm/providers"
	"github.com/openforge-ai/openforge/internal/tool"
)

func TestAgentIntegration_Ollama(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ollama := providers.NewOllamaProvider("")
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	health, err := ollama.Status(ctx)
	if err != nil || health.Status != pm.StatusAvailable {
		t.Skip("Ollama not available")
	}

	models, _ := ollama.ListModels(ctx)
	var modelID string
	for _, m := range models {
		if m.ID == "llama3.2:3b" || m.ID == "qwen2.5:3b" {
			modelID = m.ID
			break
		}
	}
	if modelID == "" && len(models) > 0 {
		modelID = models[0].ID
	}
	if modelID == "" {
		t.Skip("no models available")
	}
	t.Logf("Using model: %s", modelID)

	tools := []tool.Tool{tool.NewWriteTool()}
	ag := New(AgentConfig{
		Model: modelID, MaxTokens: 1024, Temperature: 0.7,
		Provider: ollama, Tools: tools, SystemPrompt: CoderSystemPrompt,
	})

	testFile := "agent_test_output.txt"
	os.Remove(testFile)

	result, err := ag.Run(ctx, "Write a file called agent_test_output.txt with the text 'integration test passed'", nil)
	if err != nil {
		t.Logf("Agent error: %v", err)
	}
	t.Logf("Response: %s", result)

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Logf("File not created: %v (model may have returned text instead of tool call)", err)
	} else {
		t.Logf("File content: %q", string(data))
		if string(data) == "integration test passed" {
			t.Log("✅ PASSED")
		}
	}
	os.Remove(testFile)
}

func TestAgentIntegration_Qwen(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ollama := providers.NewOllamaProvider("")
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	health, err := ollama.Status(ctx)
	if err != nil || health.Status != pm.StatusAvailable {
		t.Skip("Ollama not available")
	}

	models, _ := ollama.ListModels(ctx)
	found := false
	for _, m := range models {
		if m.ID == "qwen2.5:3b" {
			found = true
			break
		}
	}
	if !found {
		t.Skip("qwen2.5:3b not available")
	}

	tools := []tool.Tool{tool.NewWriteTool(), tool.NewViewTool()}
	ag := New(AgentConfig{
		Model: "qwen2.5:3b", MaxTokens: 1024, Temperature: 0.7,
		Provider: ollama, Tools: tools, SystemPrompt: CoderSystemPrompt,
	})

	testFile := "agent_multi_test.txt"
	os.Remove(testFile)

	result, err := ag.Run(ctx, "Create a file called agent_multi_test.txt containing 'multi-tool works', then read it back", nil)
	if err != nil {
		t.Logf("Error: %v", err)
	}
	t.Logf("Result: %s", result)

	if data, err := os.ReadFile(testFile); err == nil {
		t.Logf("Content: %q", string(data))
	}
	os.Remove(testFile)
}
