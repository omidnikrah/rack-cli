package tui

import "charm.land/bubbles/v2/key"

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Add    key.Binding
	Edit   key.Binding
	Delete key.Binding
	Ping   key.Binding
	SSH    key.Binding
	Back   key.Binding
	Filter key.Binding
	Help   key.Binding
	Reveal key.Binding
	Copy   key.Binding
	Quit   key.Binding
}

var keys = keyMap{
	Up:     key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
	Down:   key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
	Enter:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Add:    key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
	Edit:   key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	Delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	Ping:   key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "ping")),
	SSH:    key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "ssh")),
	Back:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Filter: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
	Help:   key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Reveal: key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "reveal password")),
	Copy:   key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy password")),
	Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.Enter,
		k.Back,
		k.Quit,
	}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Back},
		{k.Add, k.Edit, k.Delete},
		{k.Ping, k.SSH},
		{k.Quit},
	}
}
