package agent

import (
	"context"
	"fmt"

	"github.com/openforge-ai/openforge/internal/pm"
	"github.com/openforge-ai/openforge/internal/tool"
)

type AgentRole string

const (
	RoleCoder   AgentRole = "coder"
	RoleExplore AgentRole = "explore"
	RolePlan    AgentRole = "plan"
	RoleVerify  AgentRole = "verify"
)

var rolePrompts = map[AgentRole]string{
	RoleCoder:   "You are a coding assistant. Write, edit, and run code.",
	RoleExplore: "You are a code explorer. Search, read, and analyze code.",
	RolePlan:    "You are a planner. Analyze requirements and create plans.",
	RoleVerify:  "You are a code reviewer. Verify correctness, find bugs.",
}

type Coordinator struct {
	agents map[AgentRole]*Agent
}

func NewCoordinator(pmgr *pm.ProviderManager, tools []tool.Tool) (*Coordinator, error) {
	prov, err := pmgr.ActiveProvider(context.Background())
	if err != nil {
		return nil, fmt.Errorf("no active provider: %w", err)
	}

	c := &Coordinator{agents: make(map[AgentRole]*Agent)}
	for role, prompt := range rolePrompts {
		c.agents[role] = New(AgentConfig{
			Model: "", MaxTokens: 4096, Temperature: 0.7,
			Provider: prov, Tools: tools, SystemPrompt: prompt,
		})
	}
	return c, nil
}

func (c *Coordinator) Run(ctx context.Context, role AgentRole, prompt string, streamFn func(string)) (string, error) {
	a, ok := c.agents[role]
	if !ok {
		return "", fmt.Errorf("unknown role: %s", role)
	}
	return a.Run(ctx, prompt, streamFn)
}

func (c *Coordinator) Reset() {
	for _, a := range c.agents {
		a.Reset()
	}
}
