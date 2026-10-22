package tui

import (
	"rack-cli/internal/entity"
	"rack-cli/internal/server"
	"rack-cli/internal/storage"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

type menuScreen struct {
	list  list.Model
	stats overviewStats
}

func newMenu(servers *server.Service, domains storage.Repo[entity.Domain], ips *serverIndex) *menuScreen {
	items := []list.Item{
		menuItem{
			label: "Manage Servers", open: func() Screen {
				return newServerList(servers)
			},
		},
		menuItem{
			label: "Manage Domains", open: func() Screen {
				return newDomainList(domains, ips)
			},
		},
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false

	l := list.New(items, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)

	return &menuScreen{list: l}
}

func (ms *menuScreen) Title() string {
	return "Welcome to Rack CLI"
}

func (ms *menuScreen) Init() tea.Cmd {
	return nil
}

func (ms *menuScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ms.list.SetSize(msg.Width-4, max(msg.Height-10-overviewHeight-1, 3))
		return ms, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keys.Enter):
			if item, ok := ms.list.SelectedItem().(menuItem); ok {
				return ms, push(item.open())
			}
			return ms, nil
		}
	}

	var cmd tea.Cmd
	ms.list, cmd = ms.list.Update(msg)
	return ms, cmd
}

func (ms *menuScreen) setData(servers []entity.Server, domains []entity.Domain) {
	ms.stats = computeStats(servers, domains)
}

func (ms *menuScreen) View() string { return renderOverview(ms.stats) + "\n\n" + ms.list.View() }

type menuItem struct {
	label string
	open  func() Screen
}

func (mi menuItem) FilterValue() string {
	return mi.label
}

func (mi menuItem) Title() string {
	return mi.label
}
func (mi menuItem) Description() string {
	return ""
}
