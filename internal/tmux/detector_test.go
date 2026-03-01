package tmux_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wake/tmux-session-menu/internal/tmux"
)

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name, input, expected string
	}{
		{"plain text", "hello world", "hello world"},
		{"color codes", "\033[32mgreen\033[0m", "green"},
		{"cursor movement", "\033[2J\033[H", ""},
		{"mixed", "\033[1;33mWarning:\033[0m file not found", "Warning: file not found"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tmux.StripANSI(tt.input))
		})
	}
}

func TestDetectStatus_Running(t *testing.T) {
	tests := []struct{ name, content string }{
		{"ctrl+c prompt", "Processing files...\n  ctrl+c to interrupt"},
		{"esc prompt", "Reading code...\n  esc to interrupt"},
		{"braille spinner", "⠋ Working on task..."},
		{"asterisk spinner", "* Clauding... (12s, ↓ 200 tokens)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tmux.StatusRunning, tmux.DetectStatus(tt.content))
		})
	}
}

func TestDetectStatus_Waiting(t *testing.T) {
	tests := []struct{ name, content string }{
		{"prompt char", "Task completed.\n>"},
		{"heavy prompt", "Done.\n❯"},
		{"permission yes/no", "Allow this action?\n  Yes, allow once"},
		{"continue prompt", "Continue? (Y/n)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tmux.StatusWaiting, tmux.DetectStatus(tt.content))
		})
	}
}

func TestDetectStatus_Idle(t *testing.T) {
	tests := []struct{ name, content string }{
		{"plain shell", "user@host:~$"},
		{"empty", ""},
		{"random text", "some random output\nmore lines"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tmux.StatusIdle, tmux.DetectStatus(tt.content))
		})
	}
}
