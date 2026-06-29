package agent

import (
	"fmt"
	"os"
	"strings"

	"github.com/openforge-ai/openforge/internal/pm"
	"github.com/openforge-ai/openforge/internal/skills"
	"github.com/openforge-ai/openforge/internal/tool"
)

func BuildMessages(systemPrompt, userMessage string, history []pm.Message, toolMsgs []pm.Message) []pm.Message {
	msgs := make([]pm.Message, 0, len(history)+len(toolMsgs)+2)
	if systemPrompt != "" {
		msgs = append(msgs, pm.Message{Role: "system", Content: systemPrompt})
	}
	msgs = append(msgs, history...)
	if userMessage != "" {
		msgs = append(msgs, pm.Message{Role: "user", Content: userMessage})
	}
	msgs = append(msgs, toolMsgs...)
	return msgs
}

func BuildSystemPrompt(template string, registry *tool.Registry) string {
	wd, _ := os.Getwd()
	r := template
	r = strings.ReplaceAll(r, "{{.ToolDescriptions}}", registry.Descriptions())
	r = strings.ReplaceAll(r, "{{.WorkingDir}}", wd)

	// Auto-discover skills
	skillFiles, _ := skills.DiscoverSkills("./skills", "~/.openforge/skills")
	r = skills.InjectSkills(r, skillFiles)

	return r
}

func formatToolOutput(name, output, toolErr string) string {
	if toolErr != "" {
		return fmt.Sprintf("[%s] Error: %s\n%s", name, toolErr, output)
	}
	return fmt.Sprintf("[%s]\n%s", name, output)
}
