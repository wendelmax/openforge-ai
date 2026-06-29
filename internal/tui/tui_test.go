package tui

import (
	"testing"
)

func TestGetCompletions(t *testing.T) {
	models := []string{"phi-3-mini", "llama-3b"}
	devices := []string{"CPU", "GPU", "NPU"}

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{name: "empty", input: "", expected: []string{"/help", "/model", "/device", "/clear", "/exit", "/provider", "/tools", "/save", "/load", "/yolo"}},
		{name: "/", input: "/", expected: []string{"/help", "/model", "/device", "/clear", "/exit", "/provider", "/tools", "/save", "/load", "/yolo"}},
		{name: "/m", input: "/m", expected: []string{"/model"}},
		{name: "/model with trailing space", input: "/model ", expected: []string{"phi-3-mini", "llama-3b"}},
		{name: "/d", input: "/d", expected: []string{"/device"}},
		{name: "/device with trailing space", input: "/device ", expected: []string{"CPU", "GPU", "NPU"}},
		{name: "no match", input: "/x", expected: nil},
		{name: "plain text", input: "hello", expected: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCompletions(tt.input, models, devices)
			if len(got) != len(tt.expected) {
				t.Errorf("getCompletions(%q) = %v (len %d), want %v (len %d)",
					tt.input, got, len(got), tt.expected, len(tt.expected))
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("getCompletions(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestBuildMessageContent(t *testing.T) {
	msg := chatMessage{role: "user", content: "hello"}
	got := buildMessageContent(msg, 80, nil)
	if got == "" {
		t.Fatal("buildMessageContent returned empty")
	}
	if !containsStr(got, "hello") {
		t.Errorf("buildMessageContent should contain 'hello', got: %s", got)
	}

	msg2 := chatMessage{role: "assistant", content: "**bold** text"}
	got2 := buildMessageContent(msg2, 80, nil)
	if got2 == "" {
		t.Fatal("buildMessageContent returned empty for assistant")
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
