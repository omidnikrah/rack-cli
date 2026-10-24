package tui

import "charm.land/lipgloss/v2"

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1B1F24")).
			Background(lipgloss.Color("#E9A23B")).
			Padding(0, 1).
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

	tileStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3A444D")).
			Padding(0, 2).
			Width(17).
			MarginRight(1)
)
