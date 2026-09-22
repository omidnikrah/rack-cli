package tui

import (
	"context"
	"fmt"
	"rack-cli/internal/entity"
	"rack-cli/internal/server"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
)

type copiedMsg struct {
}

const clipboardTTL = 30 * time.Second

type serverInput struct {
	entity.Server
	Password string
}

func newServerList(svc *server.Service) *listScreen[entity.Server] {
	return newListScreen(listConfig[entity.Server]{
		title: "Manage Servers",
		id: func(s entity.Server) int {
			return s.ID
		},
		rowName: func(s entity.Server) string {
			return s.Label
		},
		rowDesc: serverRowDescription,
		filter: func(s entity.Server) string {
			return s.Label + " " + s.IP + " " + strings.Join(s.Tags, ", ")
		},
		open: func(s entity.Server) Screen {
			return newServerDetail(svc, s)
		},
		add: func() Screen {
			return newServerForm(svc, entity.Server{})
		},
		edit: func(s entity.Server) Screen {
			return newServerForm(svc, s)
		},
		remove: func(id int) tea.Cmd {
			return deleteCmd(func() error {
				return svc.Delete(id)
			})
		},
		confirm: func(s entity.Server) string {
			return fmt.Sprintf("Delete %s (%s)?", s.Label, s.IP)
		},
	})
}

func newServerForm(svc *server.Service, original entity.Server) *formScreen[serverInput] {
	return newFormScreen(formConfig[serverInput]{
		addTitle:  "Add Server",
		editTitle: "Edit Server",
		id: func(s serverInput) int {
			return s.ID
		},
		fields: []fieldSpec{
			{label: "Label"},
			{label: "IP address"},
			{label: "Root user", placeholder: "root"},
			{label: "Password", password: true},
			{label: "Purpose"},
			{label: "Tags", placeholder: "comma, separated"},
		},
		load: func(s serverInput, in []textinput.Model) {
			in[0].SetValue(s.Label)
			in[1].SetValue(s.IP)
			in[2].SetValue(s.RootUser)
			in[3].SetValue(s.Password)
			in[4].SetValue(s.Purpose)
			in[5].SetValue(strings.Join(s.Tags, ", "))
		},
		build: func(id int, in []textinput.Model) serverInput {
			root := strings.TrimSpace(in[2].Value())
			if root == "" {
				root = "root"
			}
			return serverInput{
				Server: entity.Server{
					ID:       id,
					Label:    strings.TrimSpace(in[0].Value()),
					IP:       strings.TrimSpace(in[1].Value()),
					RootUser: root,

					Purpose: strings.TrimSpace(in[4].Value()),
					Tags:    parseTags(in[5].Value()),
				},
				Password: in[3].Value(),
			}
		},
		save: func(s serverInput) tea.Cmd {
			return saveCmd(func() error {
				_, err := svc.Save(s.Server, s.Password)
				return err
			})
		},
	}, serverInput{Server: original})
}

type pingResultMsg struct {
	rtt time.Duration
	err error
}

type sshFinishedMsg struct {
	err error
}

type serverDetailScreen struct {
	svc     *server.Service
	srv     entity.Server
	plain   string
	copied  bool
	spinner spinner.Model
	pinging bool
	pinged  bool
	rtt     time.Duration
	pingErr error
}

func newServerDetail(svc *server.Service, srv entity.Server) *serverDetailScreen {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return &serverDetailScreen{svc: svc, srv: srv, spinner: sp}
}

func (s *serverDetailScreen) Title() string {
	return s.srv.Label
}

func (s *serverDetailScreen) Init() tea.Cmd {
	return nil
}

func (s *serverDetailScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case pingResultMsg:
		s.pinging, s.pinged = false, true
		s.rtt, s.pingErr = msg.rtt, msg.err
		return s, nil
	case sshFinishedMsg:
		if msg.err != nil {
			return s, fail(msg.err)
		}
		return s, nil
	case spinner.TickMsg:
		if !s.pinging {
			return s, nil
		}
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd
	case copiedMsg:
		s.copied = true
		return s, nil
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, keys.Back):
			return s, pop
		case key.Matches(msg, keys.Edit):
			return s, push(newServerForm(s.svc, s.srv))
		case key.Matches(msg, keys.Reveal):
			if s.plain != "" {
				s.plain = ""
				return s, nil
			}
			plain, err := s.svc.Password(s.srv)
			if err != nil {
				return s, fail(err)
			}
			s.plain = plain
			return s, nil
		case key.Matches(msg, keys.Copy):
			plain, err := s.svc.Password(s.srv)
			if err != nil {
				return s, fail(err)
			}
			return s, copyPasswordCmd(plain)
		case key.Matches(msg, keys.Ping):
			if s.pinging {
				return s, nil
			}
			s.pinging, s.pinged = true, false
			return s, tea.Batch(s.spinner.Tick, pingCmd(s.srv))
		case key.Matches(msg, keys.SSH):
			return s, tea.ExecProcess(server.SSHCommand(s.srv), func(err error) tea.Msg {
				return sshFinishedMsg{err: err}
			})
		}
	}

	return s, nil
}

func (s *serverDetailScreen) View() string {
	password := "••••••••"
	if s.plain != "" {
		password = s.plain
	}

	rows := [][2]string{
		{"IP", s.srv.IP},
		{"Root user", s.srv.RootUser},
		{"Password", password},
		{"Purpose", s.srv.Purpose},
		{"Tags", strings.Join(s.srv.Tags, ", ")},
	}

	var b strings.Builder
	for _, row := range rows {
		b.WriteString(labelStyle.Render(row[0]))
		b.WriteString(valueStyle.Render(row[1]))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	switch {
	case s.pinging:
		b.WriteString(s.spinner.View() + " pinging...")
	case s.pinged && s.pingErr != nil:
		b.WriteString(errorStyle.Render("unreachable: " + s.pingErr.Error()))
	case s.pinged && s.rtt > 0:
		b.WriteString(okStyle.Render("reachable in: " + s.rtt.Round(time.Millisecond).String()))
	case s.pinged:
		b.WriteString(okStyle.Render("reachable"))
	}

	if s.copied {
		b.WriteString(okStyle.Render("password copied · clears in " + clipboardTTL.String()))
		b.WriteString("\n")
	}

	return b.String()
}

func (s *serverDetailScreen) ShortHelp() []key.Binding {
	return []key.Binding{keys.Ping, keys.SSH, keys.Reveal, keys.Copy, keys.Edit, keys.Back}
}

func (s *serverDetailScreen) FullHelp() [][]key.Binding {
	return [][]key.Binding{{keys.Ping, keys.SSH, keys.Reveal, keys.Copy}, {keys.Edit, keys.Back, keys.Quit}}
}

func pingCmd(srv entity.Server) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		rtt, err := server.Ping(ctx, srv)
		return pingResultMsg{rtt: rtt, err: err}
	}
}

func serverRowDescription(s entity.Server) string {
	if len(s.Tags) == 0 {
		return fmt.Sprintf("%s · %s", s.IP, s.Purpose)
	}

	return fmt.Sprintf("%s · %s · [%s]", s.IP, s.Purpose, strings.Join(s.Tags, ", "))
}

func copyPasswordCmd(plain string) tea.Cmd {
	return tea.Sequence(
		func() tea.Msg {
			if err := clipboard.WriteAll(plain); err != nil {
				return errorMsg{err}
			}
			return copiedMsg{}
		}, tea.Tick(clipboardTTL, func(t time.Time) tea.Msg {
			if current, _ := clipboard.ReadAll(); current == plain {
				clipboard.WriteAll("")
			}
			return nil
		}))
}
