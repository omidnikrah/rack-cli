package tui

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/omidnikrah/rack-cli/internal/entity"
)

type choice struct {
	value string
	label string
	desc  string
}

func newChoicePicker(title string, choices []choice) *listScreen[choice] {
	ls := newListScreen(listConfig[choice]{
		title:   title,
		id:      func(choice) int { return 0 },
		rowName: func(c choice) string { return c.label },
		rowDesc: func(c choice) string { return c.desc },
		filter:  func(c choice) string { return c.label },
		pick: func(c choice) tea.Cmd {
			return tea.Sequence(pop, func() tea.Msg {
				return pickedMsg{value: c.value, label: c.label}
			})
		},
	})
	ls.setItems(choices)

	return ls
}

func setChoice(in []textinput.Model, vals []string, i int, value, label string) {
	vals[i] = value
	in[i].SetValue(label)
}

func serverTypeChoices() []choice {
	var choices []choice
	for _, t := range entity.ServerTypes() {
		choices = append(choices, choice{value: string(t), label: t.Label(), desc: t.Description()})
	}
	return choices
}

func billingChoices() []choice {
	var choices []choice
	for _, c := range entity.BillingCycles() {
		choices = append(choices, choice{value: string(c), label: c.Label(), desc: c.Description()})
	}
	return choices
}
