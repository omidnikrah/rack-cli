package tui

import (
	"fmt"
	"rack-cli/internal/domain"
	"rack-cli/internal/entity"
	"rack-cli/internal/storage"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

func newDomainList(repo domain.Storage) *listScreen[entity.Domain] {
	return newListScreen(listConfig[entity.Domain]{
		title: "Manage Domains",
		id: func(d entity.Domain) int {
			return d.ID
		},
		rowName: func(d entity.Domain) string {
			return d.Name
		},
		rowDesc: domainRowDescription,
		filter: func(d entity.Domain) string {
			return d.Name + " " + d.Registrar + " " + strings.Join(d.Tags, ", ")
		},
		open: func(d entity.Domain) Screen {
			return newDomainDetail(repo, d)
		},
		add: func() Screen {
			return newDomainForm(repo, entity.Domain{})
		},
		edit: func(d entity.Domain) Screen {
			return newDomainForm(repo, d)
		},
		remove: func(id int) tea.Cmd {
			return deleteCmd(func() error {
				return repo.Delete(id)
			})
		},
		confirm: func(d entity.Domain) string {
			return fmt.Sprintf("Delete %s?", d.Name)
		},
	})
}

func newDomainForm(repo domain.Storage, original entity.Domain) *formScreen[entity.Domain] {
	return newFormScreen(formConfig[entity.Domain]{
		addTitle:  "Add Domain",
		editTitle: "Edit Domain",
		id: func(d entity.Domain) int {
			return d.ID
		},
		fields: []fieldSpec{
			{label: "Name", placeholder: "example.com"},
			{label: "Server ID", placeholder: "0 = unassigned"},
			{label: "Registrar"},
			{label: "Expires At", placeholder: "2027-10-01"},
			{label: "Notes"},
			{label: "Tags", placeholder: "comma, separated"},
		},
		load: func(d entity.Domain, in []textinput.Model) {
			in[0].SetValue(d.Name)
			if d.ServerID != 0 {
				in[1].SetValue(strconv.Itoa(d.ServerID))
			}
			in[2].SetValue(d.Registrar)
			in[3].SetValue(d.ExpiresAt)
			in[4].SetValue(d.Notes)
			in[5].SetValue(strings.Join(d.Tags, ", "))
		},
		build: func(id int, in []textinput.Model) entity.Domain {
			serverID, _ := strconv.Atoi(strings.TrimSpace(in[1].Value()))
			return entity.Domain{
				ID:        id,
				Name:      strings.TrimSpace(in[0].Value()),
				ServerID:  serverID,
				Registrar: strings.TrimSpace(in[2].Value()),
				ExpiresAt: strings.TrimSpace(in[3].Value()),
				Notes:     strings.TrimSpace(in[4].Value()),
				Tags:      parseTags(in[5].Value()),
			}
		},
		save: func(d entity.Domain) tea.Cmd {
			return saveCmd(func() error {
				_, err := storage.Save(repo, d)
				return err
			})
		},
	}, original)
}

type domainDetailScreen struct {
	repo domain.Storage
	dom  entity.Domain
}

func newDomainDetail(repo domain.Storage, dom entity.Domain) *domainDetailScreen {
	return &domainDetailScreen{repo: repo, dom: dom}
}

func (d *domainDetailScreen) Title() string {
	return d.dom.Name
}

func (d *domainDetailScreen) Init() tea.Cmd {
	return nil
}

func (d *domainDetailScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok && key.Matches(k, keys.Back) {
		return d, pop
	}
	if k, ok := msg.(tea.KeyPressMsg); ok && key.Matches(k, keys.Edit) {
		return d, push(newDomainForm(d.repo, d.dom))
	}

	return d, nil
}

func (d *domainDetailScreen) View() string {
	server := "unassigned"
	if d.dom.ServerID > 0 {
		server = fmt.Sprintf("#%d", d.dom.ServerID)
	}
	rows := [][2]string{
		{"Name", d.dom.Name},
		{"Server", server},
		{"Registrar", d.dom.Registrar},
		{"Expires At", d.dom.ExpiresAt},
		{"Notes", d.dom.Notes},
		{"Tags", strings.Join(d.dom.Tags, ", ")},
	}

	var b strings.Builder
	for _, row := range rows {
		b.WriteString(labelStyle.Render(row[0]))
		b.WriteString(valueStyle.Render(row[1]))
		b.WriteString("\n")
	}

	return b.String()
}

func (d *domainDetailScreen) ShortHelp() []key.Binding { return []key.Binding{keys.Edit, keys.Back} }

func (d *domainDetailScreen) FullHelp() [][]key.Binding {
	return [][]key.Binding{{keys.Edit, keys.Back, keys.Quit}}
}

func domainRowDescription(d entity.Domain) string {
	var parts []string
	if d.Registrar != "" {
		parts = append(parts, d.Registrar)
	}
	if d.ExpiresAt != "" {
		parts = append(parts, d.ExpiresAt)
	}
	if d.ServerID != 0 {
		parts = append(parts, fmt.Sprintf("server #%d", d.ServerID))
	}
	if len(d.Tags) > 0 {
		parts = append(parts, fmt.Sprintf("[%s]", strings.Join(d.Tags, ", ")))
	}

	return strings.Join(parts, " · ")
}
