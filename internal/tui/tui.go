package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"

	"github.com/openforge-ai/openforge/internal/agent"
	"github.com/openforge-ai/openforge/internal/pm"
	"github.com/openforge-ai/openforge/internal/tool"
)

type chatMessage struct {
	role    string
	content string
}

type TUIModel struct {
	pm        *pm.ProviderManager
	agent     *agent.Agent
	p         *tea.Program
	ctx       context.Context
	cancel    context.CancelFunc

	width, height int
	viewport      viewport.Model
	input         textinput.Model
	messages      []chatMessage
	streaming     bool
	thinking      bool
	partialResp   string
	err           error

	activeDevice  string
	activeModel   string
	tokensPerSec  float64
	totalTokens   int
	inferenceTime time.Duration

	showSuggestions bool
	suggestions     []string
	suggestionIdx   int
	yoloMode        bool

	glamour *glamour.TermRenderer
}

func New(pmgr *pm.ProviderManager, provider pm.Provider, tools []tool.Tool) *TUIModel {
	ti := textinput.New()
	ti.Placeholder = "Type a message or /help"
	ti.Prompt = "❯ "
	ti.Focus()
	ti.CharLimit = 4000

	vp := viewport.New(80, 20)
	vp.Style = MainStyle
	vp.KeyMap = viewport.KeyMap{
		PageUp:   vp.KeyMap.PageUp,
		PageDown: vp.KeyMap.PageDown,
		Up:       key.Binding{},
		Down:     key.Binding{},
	}

	ctx, cancel := context.WithCancel(context.Background())

	ag := agent.New(agent.AgentConfig{
		Model: "", MaxTokens: 4096, Temperature: 0.7,
		Provider: provider, Tools: tools, SystemPrompt: agent.CoderSystemPrompt,
	})

	activeDev := "auto"
	devices, _ := provider.Status(ctx)
	if devices != nil && len(devices.Devices) > 0 { activeDev = devices.Devices[0] }

	activeModel := "auto"
	models, _ := provider.ListModels(ctx)
	if len(models) > 0 { activeModel = models[0].ID }

	gr, _ := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(80))

	return &TUIModel{
		input: ti, viewport: vp, pm: pmgr, agent: ag,
		activeDevice: activeDev, activeModel: activeModel,
		ctx: ctx, cancel: cancel, glamour: gr,
	}
}

func (m *TUIModel) SetProgram(p *tea.Program) { m.p = p }
func (m *TUIModel) Init() tea.Cmd              { return textinput.Blink }

func (m *TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width; m.height = msg.Height
		m.viewport.Width = msg.Width; m.viewport.Height = msg.Height - 4
		m.input.Width = msg.Width - 4
		m.glamour = rebuildGlamour(msg.Width-4, m.glamour != nil)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			m.cancel()
			os.MkdirAll("./data/sessions", 0755)
			m.agent.SaveSession("./data/sessions/auto.json")
			return m, tea.Quit
		case "tab":
			m.handleTab(); return m, nil
		case "enter":
			return m.handleEnter()
		case "esc":
			m.showSuggestions = false; m.suggestions = nil; return m, nil
		default:
			m.input, _ = m.input.Update(msg); m.updateSuggestions(); return m, nil
		}

	case inferenceTokenMsg:
		m.totalTokens++; m.thinking = false
		m.partialResp += string(msg)
		m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
		return m, nil

	case inferenceDoneMsg:
		m.streaming = false; m.inferenceTime = time.Duration(msg); return m, nil

	case inferenceResultMsg:
		m.streaming = false
		m.messages = append(m.messages, chatMessage{role: "assistant", content: string(msg)})
		if m.totalTokens > 0 && m.inferenceTime > 0 {
			m.tokensPerSec = float64(m.totalTokens) / m.inferenceTime.Seconds()
		}
		m.partialResp = ""
		m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
		return m, nil

	case inferenceErrorMsg:
		m.streaming = false; m.thinking = false
		m.err = fmt.Errorf("%s", string(msg)); m.partialResp = ""
		m.viewport.SetContent(m.renderMessages())
		return m, nil

	case toolCallMsg:
		m.messages = append(m.messages, chatMessage{role: "system", content: fmt.Sprintf("🔧 %s", string(msg))})
		m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *TUIModel) View() string {
	if m.width == 0 { m.width = 80 }
	if m.height == 0 { m.height = 24 }
	m.viewport.Width = m.width; m.viewport.Height = m.height - 4; m.input.Width = m.width - 4
	var b strings.Builder
	b.WriteString(RenderTitle(m.width))
	b.WriteString(TitleStyle.Render(fmt.Sprintf(" %s │ %s ", m.activeModel, m.activeDevice)))
	b.WriteString("\n")
	if len(m.messages) == 0 && !m.streaming { m.viewport.SetContent(welcomeText()) }
	b.WriteString(m.viewport.View())
	b.WriteString("\n")
	b.WriteString(InputStyle.Render(m.input.View()))
	b.WriteString("\n")
	if m.showSuggestions && len(m.suggestions) > 0 {
		b.WriteString(RenderSuggestions(m.width, m.suggestions, m.suggestionIdx)); b.WriteString("\n")
	}
	if m.thinking {
		b.WriteString(ThinkingStyle.Render(" ⠋ thinking...")); b.WriteString("\n")
	}
	b.WriteString(RenderStatusBar(m.width, m.activeDevice, m.activeModel, m.tokensPerSec, m.yoloMode))
	return MainStyle.Render(b.String())
}

func (m *TUIModel) renderMessages() string {
	var b strings.Builder
	for _, msg := range m.messages { b.WriteString(buildMessageContent(msg, m.width, m.glamour)); b.WriteString("\n\n") }
	if m.partialResp != "" {
		b.WriteString(BotLabel.Render("┃ OpenForge")); b.WriteString("\n")
		rendered := m.partialResp
		if m.glamour != nil { if r, err := m.glamour.Render(rendered); err == nil { rendered = r } }
		b.WriteString(BotMsgStyle.Render(rendered)); b.WriteString("\n\n")
	}
	if m.err != nil { b.WriteString(ErrorStyle.Render(fmt.Sprintf("✗ %v", m.err))); b.WriteString("\n\n") }
	return b.String()
}

func buildMessageContent(msg chatMessage, width int, gr *glamour.TermRenderer) string {
	var b strings.Builder
	switch msg.role {
	case "user":
		b.WriteString(UserLabel.Render("┃ You")); b.WriteString("\n"); b.WriteString(UserMsgStyle.Render(msg.content))
	case "assistant":
		b.WriteString(BotLabel.Render("┃ OpenForge")); b.WriteString("\n")
		rendered := msg.content
		if gr != nil { if r, err := gr.Render(rendered); err == nil { rendered = r } }
		b.WriteString(BotMsgStyle.Render(rendered))
	case "system":
		b.WriteString(BotLabel.Render("┃ Tool")); b.WriteString("\n"); b.WriteString(BotMsgStyle.Render(msg.content))
	}
	return b.String()
}

func welcomeText() string {
	return BotMsgStyle.Render("Welcome to OpenForge.\nType /help to see available commands, or just start typing.")
}

func rebuildGlamour(width int, had bool) *glamour.TermRenderer {
	if !had { return nil }
	gr, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(width))
	if err != nil { return nil }
	return gr
}

var knownCommands = []string{"/help", "/model", "/device", "/clear", "/exit", "/provider", "/tools", "/save", "/load", "/yolo"}

func getCompletions(input string, modelNames, deviceIDs []string) []string {
	if input == "" { return knownCommands }
	if !strings.HasPrefix(input, "/") { return nil }
	hasTrailingSpace := strings.HasSuffix(input, " ")
	trimmed := strings.TrimSpace(input)
	parts := strings.Fields(trimmed)
	if len(parts) == 0 { return knownCommands }
	if len(parts) == 1 && !hasTrailingSpace {
		cmd := parts[0]
		var matches []string
		for _, c := range knownCommands { if strings.HasPrefix(c, cmd) { matches = append(matches, c) } }
		return matches
	}
	var argPrefix string
	if len(parts) >= 2 { argPrefix = parts[1] }
	if len(parts) >= 1 {
		switch parts[0] {
		case "/model":
			var matches []string
			for _, n := range modelNames { if strings.HasPrefix(n, argPrefix) { matches = append(matches, n) } }
			return matches
		case "/device":
			var matches []string
			for _, d := range deviceIDs { if strings.HasPrefix(d, argPrefix) { matches = append(matches, d) } }
			return matches
		}
	}
	return nil
}

func (m *TUIModel) handleTab() {
	m.updateSuggestions()
	if len(m.suggestions) == 0 { return }
	if m.suggestionIdx >= len(m.suggestions) { m.suggestionIdx = 0 }
	selected := m.suggestions[m.suggestionIdx]
	m.suggestionIdx = (m.suggestionIdx + 1) % len(m.suggestions)
	input := strings.TrimSpace(m.input.Value())
	if !strings.HasPrefix(input, "/") { return }
	parts := strings.Fields(input)
	if len(parts) <= 1 { m.input.SetValue(selected + " "); m.input.SetCursor(len(m.input.Value())) } else if len(parts) == 2 { m.input.SetValue(parts[0] + " " + selected); m.input.SetCursor(len(m.input.Value())) }
	m.showSuggestions = true
}

func (m *TUIModel) updateSuggestions() {
	m.suggestions = getCompletions(m.input.Value(), m.agentModelNames(), m.deviceIDs())
	m.showSuggestions = len(m.suggestions) > 0; m.suggestionIdx = 0
}

func (m *TUIModel) modelNames() []string { return m.agentModelNames() }

func (m *TUIModel) agentModelNames() []string {
	prov, err := m.pm.ActiveProvider(m.ctx)
	if err != nil { return nil }
	models, err := prov.ListModels(m.ctx)
	if err != nil { return nil }
	names := make([]string, len(models))
	for i, mod := range models { names[i] = mod.Name }
	return names
}

func (m *TUIModel) deviceIDs() []string {
	prov, _ := m.pm.ActiveProvider(m.ctx)
	if prov == nil { return nil }
	health, _ := prov.Status(m.ctx)
	if health == nil { return nil }
	return health.Devices
}

func (m *TUIModel) handleEnter() (tea.Model, tea.Cmd) {
	input := strings.TrimSpace(m.input.Value())
	if input == "" { return m, nil }
	m.input.SetValue(""); m.showSuggestions = false; m.suggestions = nil
	if strings.HasPrefix(input, "/") { return m, m.handleCommand(input) }
	m.messages = append(m.messages, chatMessage{role: "user", content: input})
	m.streaming = true; m.thinking = true; m.partialResp = ""; m.totalTokens = 0; m.inferenceTime = 0
	m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
	return m, m.runAgentInference(input)
}

func (m *TUIModel) handleCommand(cmd string) tea.Cmd {
	args := strings.Fields(cmd)
	if len(args) == 0 { return nil }
	switch args[0] {
	case "/help", "/h":
		m.messages = append(m.messages, chatMessage{role: "assistant", content: helpText()})
	case "/tools":
		prov, _ := m.pm.ActiveProvider(m.ctx)
		if prov != nil {
			models, _ := prov.ListModels(m.ctx)
			var list []string
			for _, mod := range models { list = append(list, fmt.Sprintf("  %s (%s)", mod.ID, mod.Format)) }
			if len(list) == 0 { m.messages = append(m.messages, chatMessage{role: "assistant", content: "No models available."}) } else { m.messages = append(m.messages, chatMessage{role: "assistant", content: fmt.Sprintf("Available models:\n%s", strings.Join(list, "\n"))}) }
		}
	case "/clear", "/c":
		m.messages = nil; m.err = nil; m.agent.Reset()
	case "/exit", "/q":
		m.cancel(); return tea.Quit
	case "/provider":
		return m.showProviderStatus()
	case "/yolo":
		m.yoloMode = !m.yoloMode
		if m.yoloMode { m.messages = append(m.messages, chatMessage{role: "system", content: "🔥 YOLO mode ON"}) } else { m.messages = append(m.messages, chatMessage{role: "system", content: "🧊 YOLO mode OFF"}) }
	case "/save":
		path := "./data/sessions/auto.json"
		if len(args) > 1 { path = args[1] }
		if err := m.agent.SaveSession(path); err != nil { m.messages = append(m.messages, chatMessage{role: "assistant", content: fmt.Sprintf("Save failed: %v", err)}) } else { m.messages = append(m.messages, chatMessage{role: "assistant", content: fmt.Sprintf("Session saved to %s", path)}) }
	case "/load":
		if len(args) < 2 { m.messages = append(m.messages, chatMessage{role: "assistant", content: "Usage: /load <path>"}) } else { if err := m.agent.LoadSession(args[1]); err != nil { m.messages = append(m.messages, chatMessage{role: "assistant", content: fmt.Sprintf("Load failed: %v", err)}) } else { m.messages = nil; m.messages = append(m.messages, chatMessage{role: "assistant", content: fmt.Sprintf("Loaded session from %s", args[1])}) } }
	default:
		m.messages = append(m.messages, chatMessage{role: "assistant", content: fmt.Sprintf("Unknown command: %s. Type /help", args[0])})
	}
	return nil
}

func (m *TUIModel) showProviderStatus() tea.Cmd {
	return func() tea.Msg {
		prov, err := m.pm.ActiveProvider(m.ctx)
		if err != nil { return inferenceResultMsg(fmt.Sprintf("No active provider: %v", err)) }
		health, _ := prov.Status(m.ctx)
		if health == nil { return inferenceResultMsg(fmt.Sprintf("Provider: %s — status unknown", prov.Info().Name)) }
		return inferenceResultMsg(fmt.Sprintf("Provider: %s — %s\nDevices: %v", prov.Info().Name, health.Status, health.Devices))
	}
}

func helpText() string {
	return `OpenForge — AI Agent CLI

Commands:
  /help, /h         Show this help
  /tools            List available tools & models
  /provider         Show active provider status
  /save [path]      Save conversation to file
  /load <path>      Load conversation from file
  /yolo             Toggle YOLO mode
  /clear, /c        Clear chat
  /exit, /q         Exit

Keybindings:
  Enter             Send message
  Tab               Autocomplete
  Esc               Hide suggestions
  Ctrl+C            Exit (auto-saves session)

Just type a message and the agent will use tools to help you.`
}

type inferenceTokenMsg string
type inferenceResultMsg string
type inferenceDoneMsg time.Duration
type inferenceErrorMsg string
type toolCallMsg string

func (m *TUIModel) runAgentInference(userMessage string) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		tokenFn := func(token string) { if m.p != nil { m.p.Send(inferenceTokenMsg(token)) } }
		toolFn := func(name, args string) { if m.p != nil { m.p.Send(toolCallMsg(fmt.Sprintf("Running tool: %s(%s)", name, args))) } }
		_, err := m.agent.RunStream(m.ctx, userMessage, tokenFn, toolFn)
		if err != nil { if m.p != nil { m.p.Send(inferenceErrorMsg(err.Error())) }; return nil }
		if m.p != nil { m.p.Send(inferenceDoneMsg(time.Since(start))) }
		return nil
	}
}
