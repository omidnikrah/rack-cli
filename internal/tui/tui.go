package tui

import (
	"fmt"
	"rack-cli/internal/domain"
	"rack-cli/internal/entity"
	"rack-cli/internal/server"
	"strings"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
)

type Model struct {
	servers *server.Service
	domains domain.Storage
	ips     *serverIndex

	stack      []Screen
	serverRows []entity.Server
	domainRows []entity.Domain

	help   help.Model
	width  int
	height int
	err    error
}

func New(servers *server.Service, domains domain.Storage) *Model {
	ips := &serverIndex{}

	return &Model{
		servers: servers,
		domains: domains,
		ips:     ips,
		stack:   []Screen{newMenu(servers, domains, ips)},
		help:    help.New(),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.loadServers, m.loadDomains)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.help.SetWidth(m.width)
		return m, m.broadcast(msg)

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if c, ok := m.current().(inputCapturer); !ok || !c.capturesInput() {
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "?":
				m.help.ShowAll = !m.help.ShowAll
				return m, nil
			}
		}

	case pushScreenMsg:
		m.err = nil

		if msg.screen == nil {
			return m, nil
		}

		m.stack = append(m.stack, msg.screen)
		return m, tea.Batch(
			msg.screen.Init(),
			func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			},
			m.refreshLists(),
		)

	case popScreenMsg:
		m.err = nil
		if len(m.stack) <= 1 {
			return m, tea.Quit
		}
		m.stack = m.stack[:len(m.stack)-1]
		return m, nil

	case serversLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.serverRows = msg.servers
		m.ips.set(msg.servers)
		return m, m.refreshLists()

	case domainsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.domainRows = msg.domains
		return m, m.refreshLists()

	case savedMsg:
		updated, cmd := m.current().Update(msg)
		m.stack[len(m.stack)-1] = updated
		if msg.err == nil {
			return m, tea.Batch(cmd, m.reloadAll())
		}
		return m, cmd

	case deletedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		return m, m.reloadAll()

	case errorMsg:
		m.err = msg.err

		return m, nil
	}

	updated, cmd := m.current().Update(msg)
	m.stack[len(m.stack)-1] = updated

	return m, cmd
}

func (m *Model) View() tea.View {
	var b strings.Builder
	b.WriteString(titleStyle.Render("** " + m.current().Title() + " **"))
	b.WriteString("\n")
	b.WriteString(m.current().View())

	if m.err != nil {
		fmt.Fprintf(&b, "\n\n%s", errorStyle.Render("error: "+m.err.Error()))
	}

	b.WriteString("\n\n")

	if hk, ok := m.current().(help.KeyMap); ok {
		b.WriteString(m.help.View(hk))
	}

	v := tea.NewView(docStyle.Render(b.String()))
	v.AltScreen = true
	v.WindowTitle = "rack-cli"

	return v
}

func (m *Model) current() Screen {
	return m.stack[len(m.stack)-1]
}

func (m *Model) broadcast(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for i, s := range m.stack {
		updated, cmd := s.Update(msg)
		m.stack[i] = updated
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func (m *Model) loadServers() tea.Msg {
	rows, err := m.servers.Load()
	return serversLoadedMsg{rows, err}
}

func (m *Model) loadDomains() tea.Msg {
	rows, err := m.domains.Load()
	return domainsLoadedMsg{rows, err}
}

func (m *Model) reloadAll() tea.Cmd {
	return tea.Batch(m.loadServers, m.loadDomains)
}

func (m *Model) refreshLists() tea.Cmd {
	var cmds []tea.Cmd

	for _, sc := range m.stack {
		switch s := sc.(type) {
		case *listScreen[entity.Server]:
			cmds = append(cmds, s.setItems(m.serverRows))
		case *listScreen[entity.Domain]:
			cmds = append(cmds, s.setItems(m.domainRows))
		case *serverDetailScreen:
			for _, srv := range m.serverRows {
				if srv.ID == s.srv.ID {
					s.srv = srv
					break
				}
			}
		case *domainDetailScreen:
			for _, d := range m.domainRows {
				if d.ID == s.dom.ID {
					s.dom = d
					break
				}
			}
		}
	}

	return tea.Batch(cmds...)
}
