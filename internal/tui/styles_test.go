package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNoHardcodedWidths(t *testing.T) {
	for name, s := range map[string]lipgloss.Style{
		"MainStyle":      MainStyle,
		"StatusBarStyle": StatusBarStyle,
	} {
		if w := s.GetWidth(); w != 0 {
			t.Errorf("%s has hardcoded Width(%d); must be dynamic via m.width", name, w)
		}
	}
}
