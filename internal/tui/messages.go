package tui

import (
	"github.com/omidnikrah/rack-cli/internal/entity"

	tea "charm.land/bubbletea/v2"
)

type serversLoadedMsg struct {
	servers []entity.Server
	err     error
}

type domainsLoadedMsg struct {
	domains []entity.Domain
	err     error
}

type savedMsg struct {
	err error
}

type deletedMsg struct {
	err error
}

type errorMsg struct {
	err error
}

type pickedMsg struct {
	value string
	label string
}

func fail(err error) tea.Cmd {
	return func() tea.Msg {
		return errorMsg{err}
	}
}

func saveCmd(do func() error) tea.Cmd {
	return func() tea.Msg {
		return savedMsg{do()}
	}
}

func deleteCmd(do func() error) tea.Cmd {
	return func() tea.Msg {
		return deletedMsg{do()}
	}
}
