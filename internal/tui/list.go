package tui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

type listConfig[T any] struct {
	title   string
	id      func(T) int
	rowName func(T) string
	rowDesc func(T) string
	filter  func(T) string
	add     func() Screen
	open    func(T) Screen
	edit    func(T) Screen
	remove  func(id int) tea.Cmd
	confirm func(T) string
	pick    func(T) tea.Cmd
}

type listScreen[T any] struct {
	cfg  listConfig[T]
	list list.Model
}

func newListScreen[T any](cfg listConfig[T]) *listScreen[T] {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)

	return &listScreen[T]{cfg: cfg, list: l}
}

func (ls *listScreen[T]) Title() string {
	return ls.cfg.title
}

func (ls *listScreen[T]) Init() tea.Cmd {
	return nil
}

func (ls *listScreen[T]) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ls.list.SetSize(msg.Width-4, msg.Height-10)
		return ls, nil

	case tea.KeyPressMsg:
		if ls.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, keys.Back):
			return ls, pop
		case key.Matches(msg, keys.Add):
			if ls.cfg.pick != nil {
				return ls, nil
			}
			return ls, push(ls.cfg.add())
		case key.Matches(msg, keys.Enter):
			if row, ok := ls.selected(); ok {
				if ls.cfg.pick != nil {
					return ls, ls.cfg.pick(row)
				}
				return ls, push(ls.cfg.open(row))
			}
			return ls, nil
		case key.Matches(msg, keys.Edit) && ls.cfg.pick == nil:
			if row, ok := ls.selected(); ok {
				return ls, push(ls.cfg.edit(row))
			}
		case key.Matches(msg, keys.Delete) && ls.cfg.pick == nil:
			if row, ok := ls.selected(); ok {
				return ls, push(newConfirm(ls.cfg.confirm(row), ls.cfg.remove(ls.cfg.id(row))))
			}
		}
	}

	var cmd tea.Cmd
	ls.list, cmd = ls.list.Update(msg)
	return ls, cmd
}

func (ls *listScreen[T]) View() string { return ls.list.View() }

func (ls *listScreen[T]) ShortHelp() []key.Binding {
	if ls.cfg.pick != nil {
		return []key.Binding{keys.Up, keys.Down, keys.Enter, keys.Filter, keys.Back}
	}

	return []key.Binding{
		keys.Up,
		keys.Down,
		keys.Enter,
		keys.Filter,
		keys.Add,
		keys.Edit,
		keys.Delete,
		keys.Back,
	}
}

func (ls *listScreen[T]) FullHelp() [][]key.Binding {
	if ls.cfg.pick != nil {
		return [][]key.Binding{{keys.Up, keys.Down, keys.Enter, keys.Filter}, {keys.Back, keys.Quit, keys.Help}}
	}

	return [][]key.Binding{
		{keys.Up, keys.Down, keys.Enter},
		{keys.Add, keys.Edit, keys.Delete, keys.Filter},
		{keys.Back, keys.Quit, keys.Help},
	}
}

func (ls *listScreen[T]) capturesInput() bool {
	return ls.list.FilterState() == list.Filtering
}

func (ls *listScreen[T]) setItems(items []T) tea.Cmd {
	rows := make([]list.Item, len(items))

	for i, item := range items {
		rows[i] = listItem[T]{v: item, cfg: ls.cfg}
	}

	return ls.list.SetItems(rows)
}

func (ls *listScreen[T]) selected() (T, bool) {
	item, ok := ls.list.SelectedItem().(listItem[T])
	return item.v, ok
}

type listItem[T any] struct {
	v   T
	cfg listConfig[T]
}

func (li listItem[T]) FilterValue() string {
	return li.cfg.filter(li.v)
}

func (li listItem[T]) Title() string {
	return li.cfg.rowName(li.v)
}

func (li listItem[T]) Description() string {
	return li.cfg.rowDesc(li.v)
}
