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

func TestModel_Navigation(t *testing.T) {
	m := ui.NewModel()
	m.SetItems([]ui.ListItem{
		{Type: ui.ItemSession},
		{Type: ui.ItemSession},
		{Type: ui.ItemSession},
	})

	assert.Equal(t, 0, m.Cursor())

	m, _ = applyKey(m, "j")
	assert.Equal(t, 1, m.Cursor())

	m, _ = applyKey(m, "k")
	assert.Equal(t, 0, m.Cursor())

	// Can't go past top
	m, _ = applyKey(m, "k")
	assert.Equal(t, 0, m.Cursor())

	// Can't go past bottom
	m, _ = applyKey(m, "j")
	m, _ = applyKey(m, "j")
	m, _ = applyKey(m, "j")
	assert.Equal(t, 2, m.Cursor())
}

func applyKey(m ui.Model, key string) (ui.Model, tea.Cmd) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	updated, cmd := m.Update(msg)
	return updated.(ui.Model), cmd
}
