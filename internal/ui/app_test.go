package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/wake/tmux-session-menu/internal/ui"
)

func TestModel_Init(t *testing.T) {
	m := ui.NewModel()
	cmd := m.Init()
	assert.NotNil(t, cmd)
}

func TestModel_View_ShowsHeader(t *testing.T) {
	m := ui.NewModel()
	view := m.View()
	assert.Contains(t, view, "tmux session menu")
}

func TestModel_Quit(t *testing.T) {
	m := ui.NewModel()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	_ = updated
	assert.NotNil(t, cmd)
}
