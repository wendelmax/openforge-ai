package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/openforge-ai/openforge/runtime"
)

type chatMessage struct {
	role    string
	content string
}

type TUIModel struct {
	rt      runtime.Runtime
	p       *tea.Program
	ctx     context.Context
	cancel  context.CancelFunc

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
	devices       []runtime.Device
	models        []runtime.ModelInfo

	showSuggestions bool
	suggestions     []string
	suggestionIdx   int

	glamour *glamour.TermRenderer
}

func New(rt runtime.Runtime) *TUIModel {
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

	devices, _ := rt.ListDevices(ctx)
	models, _ := rt.ListModels(ctx)

	activeDev := "CPU"
	for _, d := range devices {
		if d.Available {
			activeDev = d.ID
			break
		}
	}

	activeModel := "none"
	if len(models) > 0 {
		activeModel = models[0].ID
	}

	gr, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

	return &TUIModel{
		input:         ti,
		viewport:      vp,
		rt:            rt,
		activeDevice:  activeDev,
		activeModel:   activeModel,
		devices:       devices,
		models:        models,
		ctx:           ctx,
		cancel:        cancel,
		glamour:       gr,
	}
}

func (m *TUIModel) SetProgram(p *tea.Program) { m.p = p }

func (m *TUIModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 4
		m.input.Width = msg.Width - 4
		m.glamour = rebuildGlamour(msg.Width-4, m.glamour != nil)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			m.cancel()
			return m, tea.Quit

		case "tab":
			m.handleTab()
			return m, nil

		case "enter":
			return m.handleEnter()

		case "esc":
			m.showSuggestions = false
			m.suggestions = nil
			return m, nil

		default:
			m.input, _ = m.input.Update(msg)
			m.updateSuggestions()
			return m, nil
		}

	case inferenceTokenMsg:
		m.totalTokens++
		m.thinking = false
		m.partialResp += string(msg)
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case inferenceDoneMsg:
		m.streaming = false
		m.inferenceTime = time.Duration(msg)
		return m, nil

	case inferenceResultMsg:
		m.streaming = false
		m.messages = append(m.messages, chatMessage{role: "assistant", content: string(msg)})
		if m.totalTokens > 0 && m.inferenceTime > 0 {
			m.tokensPerSec = float64(m.totalTokens) / m.inferenceTime.Seconds()
		}
		m.partialResp = ""
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case inferenceErrorMsg:
		m.streaming = false
		m.thinking = false
		m.err = fmt.Errorf("%s", string(msg))
		m.partialResp = ""
		m.viewport.SetContent(m.renderMessages())
		return m, nil

	case modelsUpdatedMsg:
		m.models = []runtime.ModelInfo(msg)
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *TUIModel) View() string {
	if m.width == 0 {
		m.width = 80
	}
	if m.height == 0 {
		m.height = 24
	}

	m.viewport.Width = m.width
	m.viewport.Height = m.height - 4
	m.input.Width = m.width - 4

	var b strings.Builder

	b.WriteString(RenderTitle(m.width))
	b.WriteString(TitleStyle.Render(fmt.Sprintf(" %s │ %s ", m.activeModel, m.activeDevice)))
	b.WriteString("\n")

	if len(m.messages) == 0 && !m.streaming {
		m.viewport.SetContent(welcomeText())
	}
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	b.WriteString(InputStyle.Render(m.input.View()))
	b.WriteString("\n")

	if m.showSuggestions && len(m.suggestions) > 0 {
		b.WriteString(RenderSuggestions(m.width, m.suggestions, m.suggestionIdx))
		b.WriteString("\n")
	}

	if m.thinking {
		b.WriteString(ThinkingStyle.Render(" ⠋ thinking..."))
		b.WriteString("\n")
	}

	b.WriteString(RenderStatusBar(m.width, m.activeDevice, m.activeModel, m.tokensPerSec))

	return MainStyle.Render(b.String())
}

func (m *TUIModel) renderMessages() string {
	var b strings.Builder

	for _, msg := range m.messages {
		b.WriteString(buildMessageContent(msg, m.width, m.glamour))
		b.WriteString("\n\n")
	}

	if m.partialResp != "" {
		b.WriteString(BotLabel.Render("┃ OpenForge"))
		b.WriteString("\n")
		rendered := m.partialResp
		if m.glamour != nil {
			if r, err := m.glamour.Render(rendered); err == nil {
				rendered = r
			}
		}
		b.WriteString(BotMsgStyle.Render(rendered))
		b.WriteString("\n\n")
	}

	if m.err != nil {
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("✗ %v", m.err)))
		b.WriteString("\n\n")
	}

	return b.String()
}

func buildMessageContent(msg chatMessage, width int, gr *glamour.TermRenderer) string {
	var b strings.Builder

	switch msg.role {
	case "user":
		b.WriteString(UserLabel.Render("┃ You"))
		b.WriteString("\n")
		b.WriteString(UserMsgStyle.Render(msg.content))
	case "assistant":
		b.WriteString(BotLabel.Render("┃ OpenForge"))
		b.WriteString("\n")
		rendered := msg.content
		if gr != nil {
			if r, err := gr.Render(rendered); err == nil {
				rendered = r
			}
		}
		b.WriteString(BotMsgStyle.Render(rendered))
	}

	return b.String()
}

func welcomeText() string {
	return BotMsgStyle.Render(
		"Welcome to OpenForge.\nType /help to see available commands, or just start typing.",
	)
}

func rebuildGlamour(width int, had bool) *glamour.TermRenderer {
	if !had {
		return nil
	}
	gr, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}
	return gr
}

// ── Completion engine ─────────────────────────────────────────────

var knownCommands = []string{"/help", "/model", "/device", "/clear", "/exit"}

func getCompletions(input string, modelNames, deviceIDs []string) []string {
	if input == "" {
		return knownCommands
	}

	if !strings.HasPrefix(input, "/") {
		return nil
	}

	// Check for trailing space — user wants argument completion
	hasTrailingSpace := strings.HasSuffix(input, " ")

	trimmed := strings.TrimSpace(input)
	parts := strings.Fields(trimmed)

	if len(parts) == 0 {
		return knownCommands
	}

	if len(parts) == 1 && !hasTrailingSpace {
		cmd := parts[0]
		var matches []string
		for _, c := range knownCommands {
			if strings.HasPrefix(c, cmd) {
				matches = append(matches, c)
			}
		}
		return matches
	}

	var argPrefix string
	if len(parts) >= 2 {
		argPrefix = parts[1]
	} else if hasTrailingSpace {
		argPrefix = ""
	}

	if len(parts) >= 1 {
		switch parts[0] {
		case "/model":
			var matches []string
			for _, n := range modelNames {
				if strings.HasPrefix(n, argPrefix) {
					matches = append(matches, n)
				}
			}
			return matches
		case "/device":
			var matches []string
			for _, d := range deviceIDs {
				if strings.HasPrefix(d, argPrefix) {
					matches = append(matches, d)
				}
			}
			return matches
		}
	}

	return nil
}

func (m *TUIModel) handleTab() {
	m.updateSuggestions()

	if len(m.suggestions) == 0 {
		return
	}

	if m.suggestionIdx >= len(m.suggestions) {
		m.suggestionIdx = 0
	}

	selected := m.suggestions[m.suggestionIdx]
	m.suggestionIdx = (m.suggestionIdx + 1) % len(m.suggestions)

	input := strings.TrimSpace(m.input.Value())

	if !strings.HasPrefix(input, "/") {
		return
	}

	parts := strings.Fields(input)

	if len(parts) <= 1 {
		m.input.SetValue(selected + " ")
		m.input.SetCursor(len(m.input.Value()))
	} else if len(parts) == 2 {
		m.input.SetValue(parts[0] + " " + selected)
		m.input.SetCursor(len(m.input.Value()))
	}

	m.showSuggestions = true
}

func (m *TUIModel) updateSuggestions() {
	m.suggestions = getCompletions(m.input.Value(), m.modelNames(), m.deviceIDs())
	m.showSuggestions = len(m.suggestions) > 0
	m.suggestionIdx = 0
}

func (m *TUIModel) modelNames() []string {
	names := make([]string, len(m.models))
	for i, mod := range m.models {
		names[i] = mod.Name
	}
	return names
}

func (m *TUIModel) deviceIDs() []string {
	ids := make([]string, len(m.devices))
	for i, d := range m.devices {
		ids[i] = d.ID
	}
	return ids
}

// ── Commands ────────────────────────────────────────────────────

func (m *TUIModel) handleEnter() (tea.Model, tea.Cmd) {
	input := strings.TrimSpace(m.input.Value())
	if input == "" {
		return m, nil
	}
	m.input.SetValue("")
	m.showSuggestions = false
	m.suggestions = nil

	if strings.HasPrefix(input, "/") {
		return m, m.handleCommand(input)
	}

	if m.activeModel == "none" || m.activeModel == "" {
		m.err = fmt.Errorf("no model loaded. Use /model <name> to load one")
		return m, nil
	}

	m.messages = append(m.messages, chatMessage{role: "user", content: input})
	m.streaming = true
	m.thinking = true
	m.partialResp = ""
	m.totalTokens = 0
	m.inferenceTime = 0
	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()

	return m, m.runInference(input)
}

func (m *TUIModel) handleCommand(cmd string) tea.Cmd {
	args := strings.Fields(cmd)
	if len(args) == 0 {
		return nil
	}

	switch args[0] {
	case "/help", "/h":
		m.messages = append(m.messages, chatMessage{role: "assistant", content: helpText()})

	case "/model", "/m":
		if len(args) > 1 {
			name := args[1]
			m.activeModel = name
			err := m.rt.LoadModel(m.ctx, name, name, m.activeDevice)
			if err != nil {
				m.err = fmt.Errorf("load model: %w", err)
			} else {
				m.messages = append(m.messages, chatMessage{role: "assistant", content: fmt.Sprintf("Switched to model: %s", name)})
			}
		} else {
			var list []string
			for _, mod := range m.models {
				status := ""
				if mod.Loaded {
					status = " [loaded]"
				}
				list = append(list, fmt.Sprintf("  %s%s", mod.Name, status))
			}
			if len(list) == 0 {
				m.messages = append(m.messages, chatMessage{role: "assistant", content: "No models found in model directory."})
			} else {
				m.messages = append(m.messages, chatMessage{role: "assistant", content: fmt.Sprintf("Available models:\n%s", strings.Join(list, "\n"))})
			}
		}

	case "/device", "/d":
		if len(args) > 1 {
			m.activeDevice = args[1]
			m.messages = append(m.messages, chatMessage{role: "assistant", content: fmt.Sprintf("Switched to device: %s", args[1])})
		} else {
			var list []string
			for _, d := range m.devices {
				list = append(list, fmt.Sprintf("  %s", d.ID))
			}
			m.messages = append(m.messages, chatMessage{role: "assistant", content: fmt.Sprintf("Available devices:\n%s", strings.Join(list, "\n"))})
		}

	case "/clear", "/c":
		m.messages = nil
		m.err = nil

	case "/models":
		return m.reloadModels()

	case "/exit", "/q":
		m.cancel()
		return tea.Quit

	default:
		m.messages = append(m.messages, chatMessage{role: "assistant", content: fmt.Sprintf("Unknown command: %s. Type /help", args[0])})
	}
	return nil
}

func (m *TUIModel) reloadModels() tea.Cmd {
	return func() tea.Msg {
		models, err := m.rt.ListModels(m.ctx)
		if err != nil {
			return inferenceErrorMsg(err.Error())
		}
		return modelsUpdatedMsg(models)
	}
}

func helpText() string {
	return `OpenForge — AI Runtime CLI

Commands:
  /help, /h        Show this help
  /model, /m       List available models
  /model <name>    Load and switch to a model
  /device, /d      List available devices
  /device <name>   Switch to device
  /clear, /c       Clear chat
  /exit, /q        Exit

Keybindings:
  Enter            Send message
  Tab              Autocomplete commands / model names / device IDs
  Esc              Hide suggestions
  PgUp / PgDn      Scroll chat history
  Ctrl+C           Exit

Tab completes:
  • /commands — type / then Tab
  • model names — after /model 
  • device IDs — after /device

Just type a message and press Enter to chat.
Use /model <name> to load a model first.`
}

// ── Streaming ───────────────────────────────────────────────────

type inferenceTokenMsg string
type inferenceResultMsg string
type inferenceDoneMsg time.Duration
type inferenceErrorMsg string
type modelsUpdatedMsg []runtime.ModelInfo

func (m *TUIModel) runInference(prompt string) tea.Cmd {
	return func() tea.Msg {
		ch, err := m.rt.InferStream(m.ctx, m.activeModel, prompt, runtime.InferenceParams{
			Device:    m.activeDevice,
			MaxTokens: 1024,
		})
		if err != nil {
			return inferenceErrorMsg(err.Error())
		}

		start := time.Now()
		go func() {
			var fullText strings.Builder
			for token := range ch {
				fullText.WriteString(token)
				if m.p != nil {
					m.p.Send(inferenceTokenMsg(token))
				}
			}
			m.inferenceTime = time.Since(start)
			if m.p != nil {
				if fullText.Len() == 0 {
					m.p.Send(inferenceResultMsg("(empty response)"))
				} else {
					m.p.Send(inferenceResultMsg(fullText.String()))
				}
			}
		}()

		return nil
	}
}
