package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Model 是 Bubble Tea 的主要模型。
type Model struct {
	width    int
	height   int
	cursor   int
	items    []ListItem
	quitting bool
}

// NewModel 建立初始 Model。
func NewModel() Model {
	return Model{}
}

// Init 實作 tea.Model 介面。
func (m Model) Init() tea.Cmd {
	return tea.SetWindowTitle("tmux session menu")
}

// Update 處理訊息並更新模型狀態。
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "j", "down":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
			return m, nil
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		}
	}
	return m, nil
}

// View 渲染 TUI 畫面。
func (m Model) View() string {
	if m.quitting {
		return ""
	}
	header := headerStyle.Render("tmux session menu") +
		dimStyle.Render("  (↑↓/jk 選擇, Enter 確認, q 離開)")
	help := fmt.Sprintf("\n  %s  %s  %s",
		dimStyle.Render("[n] 新建"),
		dimStyle.Render("[g] 新群組"),
		dimStyle.Render("[q] 離開"))
	return header + "\n" + help + "\n"
}

// SetItems 設定列表項目（主要用於測試）。
func (m *Model) SetItems(items []ListItem) {
	m.items = items
	if m.cursor >= len(items) && len(items) > 0 {
		m.cursor = len(items) - 1
	}
}

// Cursor 回傳目前游標位置。
func (m Model) Cursor() int {
	return m.cursor
}
