package tui

import (
	tea "charm.land/bubbletea/v2"
)

type Screen interface {
	Title() string
	Init() tea.Cmd
	Update(tea.Msg) (Screen, tea.Cmd)
	View() string
}

type inputCapturer interface {
	capturesInput() bool
}

type pushScreenMsg struct {
	screen Screen
}

type popScreenMsg struct{}

func push(s Screen) tea.Cmd {
	return func() tea.Msg {
		return pushScreenMsg{screen: s}
	}
}

func pop() tea.Msg {
	return popScreenMsg{}
}
