package tui

import "charm.land/lipgloss/v2"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#E9A23B")).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8593A0")).
			Width(12)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E4E8EC"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D9784E"))

	okStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#5FAE83"))

	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E9A23B"))

	blurredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8593A0"))

	docStyle = lipgloss.NewStyle().Padding(1, 2)
)
