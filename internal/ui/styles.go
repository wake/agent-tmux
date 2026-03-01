package ui

import "github.com/charmbracelet/lipgloss"

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#c0caf5"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7")).Bold(true)
	dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#787fa0"))
	statusRunningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a"))
	statusWaitingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#e0af68"))
	statusIdleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#787fa0"))
	statusErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e"))
	previewBorderStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color("#414868"))
)
