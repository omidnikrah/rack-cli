package tui

import (
	"context"
	"errors"
	"fmt"
	"github.com/omidnikrah/rack-cli/internal/entity"
	"github.com/omidnikrah/rack-cli/internal/server"
	"strconv"
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

var errOneTimeRenew = errors.New("one-time servers cannot be renewed")

const (
	serverLabelField = iota
	serverTypeField
	serverIPField
	serverUserField
	serverPasswordField
	serverProviderField
	serverBillingField
	serverPriceField
	serverExpiredField
	serverPurposeField
	serverTagsField
)

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
			return s.Label + " " + s.IP + " " + s.Provider + " " + strings.Join(s.Tags, ", ")
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
			return actionCmd(func() error {
				return svc.Delete(id)
			})
		},
		confirm: func(s entity.Server) string {
			if !s.HasHost() {
				return fmt.Sprintf("Delete %s?", s.Label)
			}
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
			serverLabelField: {label: "Label"},
			serverTypeField: {label: "Type", placeholder: "enter to pick", required: true, pick: func() Screen {
				return newChoicePicker("Select Type", serverTypeChoices())
			}},
			serverIPField:       {label: "IP address", hidden: hideRootAccessFields},
			serverUserField:     {label: "Login user", placeholder: "root", hidden: hideRootAccessFields},
			serverPasswordField: {label: "Password", password: true, hidden: hideRootAccessFields},
			serverProviderField: {label: "Provider", placeholder: "Hetzner"},
			serverBillingField: {label: "Billing", placeholder: "enter to pick", required: true, pick: func() Screen {
				return newChoicePicker("Select Billing Cycle", billingChoices())
			}},
			serverPriceField:   {label: "Price", placeholder: "in USD, per billing cycle"},
			serverExpiredField: {label: "Expired At", placeholder: "2026-12-31"},
			serverPurposeField: {label: "Purpose"},
			serverTagsField:    {label: "Tags", placeholder: "comma, separated"},
		},
		load: func(s serverInput, in []textinput.Model, vals []string) {
			kind, cycle := s.Type.Normalize(), s.Billing.Normalize()

			in[serverLabelField].SetValue(s.Label)
			setChoice(in, vals, serverTypeField, string(kind), kind.Label())
			in[serverIPField].SetValue(s.IP)
			in[serverUserField].SetValue(s.RootUser)
			in[serverPasswordField].SetValue(s.Password)
			in[serverProviderField].SetValue(s.Provider)
			setChoice(in, vals, serverBillingField, string(cycle), cycle.Label())
			in[serverPriceField].SetValue(pricePtrToString(s.Price))
			in[serverExpiredField].SetValue(s.ExpiredAt)
			in[serverPurposeField].SetValue(s.Purpose)
			in[serverTagsField].SetValue(strings.Join(s.Tags, ", "))
		},
		build: func(id int, in []textinput.Model, vals []string) serverInput {
			kind := entity.ServerType(vals[serverTypeField]).Normalize()

			ip, root, password := "", "", ""
			if kind.HasRootAccess() {
				ip = strings.TrimSpace(in[serverIPField].Value())
				root = strings.TrimSpace(in[serverUserField].Value())
				password = in[serverPasswordField].Value()
				if root == "" {
					root = "root"
				}
			}

			price, _ := toPricePtr(in[serverPriceField].Value())

			return serverInput{
				Server: entity.Server{
					ID:        id,
					Label:     strings.TrimSpace(in[serverLabelField].Value()),
					Type:      kind,
					IP:        ip,
					RootUser:  root,
					Provider:  strings.TrimSpace(in[serverProviderField].Value()),
					Billing:   entity.BillingCycle(vals[serverBillingField]).Normalize(),
					Price:     price,
					ExpiredAt: strings.TrimSpace(in[serverExpiredField].Value()),
					Purpose:   strings.TrimSpace(in[serverPurposeField].Value()),
					Tags:      parseTags(in[serverTagsField].Value()),
				},
				Password: password,
			}
		},
		validate: func(s serverInput) error {
			return s.Validate()
		},
		save: func(s serverInput) tea.Cmd {
			return saveCmd(func() error {
				_, err := svc.Save(s.Server, s.Password)
				return err
			})
		},
	}, serverInput{Server: original})
}

func newServerPicker() *listScreen[entity.Server] {
	return newListScreen(listConfig[entity.Server]{
		title: "Select Server",
		id: func(s entity.Server) int {
			return s.ID
		},
		rowName: func(s entity.Server) string {
			return s.Label
		},
		rowDesc: serverRowDescription,
		filter: func(s entity.Server) string {
			return s.Label + " " + s.IP + " " + s.Provider + " " + strings.Join(s.Tags, ", ")
		},
		pick: func(s entity.Server) tea.Cmd {
			return tea.Sequence(pop, func() tea.Msg {
				return pickedMsg{value: strconv.Itoa(s.ID), label: hostLabel(s)}
			})
		},
	})
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

func (s *serverDetailScreen) rootAccess() bool {
	return s.srv.Type.HasRootAccess()
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
		case key.Matches(msg, keys.Renew):
			if !s.srv.CanRenew() {
				return s, fail(errOneTimeRenew)
			}
			return s, renewConfirm(s.srv.Label, s.srv.Renewed(time.Now()).ExpiredAt, renewServerCmd(s.svc, s.srv.ID))
		case key.Matches(msg, keys.Reveal):
			if !s.rootAccess() {
				return s, nil
			}
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
			if !s.rootAccess() {
				return s, nil
			}
			plain, err := s.svc.Password(s.srv)
			if err != nil {
				return s, fail(err)
			}
			return s, copyPasswordCmd(plain)
		case key.Matches(msg, keys.Ping):
			if s.pinging || !s.rootAccess() {
				return s, nil
			}
			s.pinging, s.pinged = true, false
			return s, tea.Batch(s.spinner.Tick, pingCmd(s.srv))
		case key.Matches(msg, keys.SSH):
			if !s.rootAccess() {
				return s, nil
			}
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

	rows := [][2]string{{"Type", s.srv.Type.Label()}}
	if s.srv.HasHost() {
		rows = append(rows, [2]string{"IP", s.srv.IP})
	}
	if s.srv.RootUser != "" {
		rows = append(rows, [2]string{"Login user", s.srv.RootUser})
	}
	if s.rootAccess() {
		rows = append(rows, [2]string{"Password", password})
	}
	rows = append(rows,
		[2]string{"Provider", s.srv.Provider},
		[2]string{"Billing", s.srv.Billing.Label()},
		[2]string{"Price", priceWithCycle(s.srv)},
		[2]string{"Expired at", s.srv.ExpiredAt},
		[2]string{"Purpose", s.srv.Purpose},
		[2]string{"Tags", strings.Join(s.srv.Tags, ", ")},
	)

	days, ok := s.srv.DaysLeft()

	var b strings.Builder
	b.WriteString(renderRows(rows, map[string]string{"Expired at": expiryBadge(days, ok)}))
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
	var bindings []key.Binding
	if s.rootAccess() {
		bindings = append(bindings, keys.Ping, keys.SSH, keys.Reveal, keys.Copy)
	}
	if s.srv.CanRenew() {
		bindings = append(bindings, keys.Renew)
	}

	return append(bindings, keys.Edit, keys.Back)
}

func (s *serverDetailScreen) FullHelp() [][]key.Binding {
	var groups [][]key.Binding
	if s.rootAccess() {
		groups = append(groups, []key.Binding{keys.Ping, keys.SSH, keys.Reveal, keys.Copy})
	}

	manage := []key.Binding{keys.Edit, keys.Back, keys.Quit}
	if s.srv.CanRenew() {
		manage = append([]key.Binding{keys.Renew}, manage...)
	}

	return append(groups, manage)
}

func renewServerCmd(svc *server.Service, id int) tea.Cmd {
	return actionCmd(func() error {
		current, err := svc.Get(id)
		if err != nil {
			return err
		}
		if !current.CanRenew() {
			return errOneTimeRenew
		}

		return svc.Update(current.Renewed(time.Now()))
	})
}

func pingCmd(srv entity.Server) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		rtt, err := server.Ping(ctx, srv)
		return pingResultMsg{rtt: rtt, err: err}
	}
}

func hideRootAccessFields(vals []string) bool {
	return !entity.ServerType(vals[serverTypeField]).HasRootAccess()
}

func serverRowDescription(s entity.Server) string {
	parts := []string{s.Type.Label()}
	for _, p := range []string{s.IP, s.Provider, s.Purpose} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	if len(s.Tags) > 0 {
		parts = append(parts, fmt.Sprintf("[%s]", strings.Join(s.Tags, ", ")))
	}

	days, ok := s.DaysLeft()
	return withBadge(strings.Join(parts, " · "), expiryBadge(days, ok))
}

func hostLabel(s entity.Server) string {
	if s.HasHost() {
		return s.IP
	}
	return s.Label
}

func priceWithCycle(s entity.Server) string {
	if s.Price == nil {
		return ""
	}
	return pricePtrToString(s.Price) + " " + s.Billing.PerLabel()
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
