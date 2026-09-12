package tui

import (
	tea "charm.land/bubbletea/v2"
)

type confirmScreen struct {
	prompt string
	onYes  tea.Cmd
}

func newConfirm(prompt string, onYes tea.Cmd) *confirmScreen {
	return &confirmScreen{prompt: prompt, onYes: onYes}
}

func (s *confirmScreen) Title() string {
	return "Confirm"
}

func (s *confirmScreen) Init() tea.Cmd {
	return nil
}

func (s *confirmScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}
	switch key.String() {
	case "y":
		if s.onYes == nil {
			return s, pop
		}
		return s, tea.Batch(s.onYes, pop)
	case "n", "esc":
		return s, pop
	}

	return s, nil
}

func (s *confirmScreen) View() string {
	return s.prompt + "\n\n" + errorStyle.Render("y") + " yes · " + okStyle.Render("n") + " cancel"
}
