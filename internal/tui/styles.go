package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	colPrimary   = lipgloss.Color("#7C3AED")
	colSecondary = lipgloss.Color("#06B6D4")
	colSuccess   = lipgloss.Color("#22C55E")
	colWarning   = lipgloss.Color("#F59E0B")
	colError     = lipgloss.Color("#EF4444")
	colText      = lipgloss.Color("#E2E8F0")
	colDim       = lipgloss.Color("#64748B")
	colBg        = lipgloss.Color("#0F172A")
	colBgAlt     = lipgloss.Color("#1E293B")

	MainStyle = lipgloss.NewStyle().
			Background(colBg)

	TitleStyle = lipgloss.NewStyle().
			Foreground(colPrimary).
			Bold(true)

	UserLabel = lipgloss.NewStyle().
			Foreground(colPrimary).
			Bold(true)

	BotLabel = lipgloss.NewStyle().
			Foreground(colSecondary).
			Bold(true)

	UserMsgStyle = lipgloss.NewStyle().
			Foreground(colText).
			PaddingLeft(2)

	BotMsgStyle = lipgloss.NewStyle().
			Foreground(colText).
			PaddingLeft(2)

	InputStyle = lipgloss.NewStyle().
			Foreground(colText).
			Background(colBgAlt).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(colError).
			Bold(true).
			PaddingLeft(1)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(colDim)

	StatusValue = lipgloss.NewStyle().
			Foreground(colSuccess)

	SuggestionStyle = lipgloss.NewStyle().
			Foreground(colDim).
			PaddingLeft(2)

	SuggestionHighlight = lipgloss.NewStyle().
				Foreground(colSecondary).
				Bold(true).
				PaddingLeft(2)

	ThinkingStyle = lipgloss.NewStyle().
			Foreground(colDim).
			PaddingLeft(2)
)

func RenderTitle(width int) string {
	return TitleStyle.Render(" ◆ OpenForge ")
}

func RenderStatusBar(width int, device, model string, tokensPerSec float64) string {
	dev := StatusValue.Render(device)
	mod := StatusValue.Render(model)

	tps := "—"
	if tokensPerSec > 0 {
		tps = StatusValue.Render(fmt.Sprintf("%.1f tok/s", tokensPerSec))
	}

	left := fmt.Sprintf(" %s │ %s │ %s ", dev, mod, tps)
	right := " /help • Ctrl+C "

	spacing := width - lipgloss.Width(left) - lipgloss.Width(right)
	if spacing < 1 {
		spacing = 1
	}

	return StatusBarStyle.Render(left + strings.Repeat(" ", spacing) + right)
}

func RenderSuggestions(width int, suggestions []string, selected int) string {
	if len(suggestions) == 0 {
		return ""
	}
	var parts []string
	for i, s := range suggestions {
		if i == selected {
			parts = append(parts, SuggestionHighlight.Render(s))
		} else {
			parts = append(parts, SuggestionStyle.Render(s))
		}
	}
	return strings.Join(parts, "  ")
}
