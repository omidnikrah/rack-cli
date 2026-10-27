package tui

import (
	"fmt"
	"github.com/omidnikrah/rack-cli/internal/entity"
	"github.com/omidnikrah/rack-cli/internal/storage"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

const serverField = 1

func newDomainList(repo storage.Repo[entity.Domain], ips *serverIndex) *listScreen[entity.Domain] {
	return newListScreen(listConfig[entity.Domain]{
		title: "Manage Domains",
		id: func(d entity.Domain) int {
			return d.ID
		},
		rowName: func(d entity.Domain) string {
			return d.Name
		},
		rowDesc: func(d entity.Domain) string {
			return domainRowDescription(d, ips)
		},
		filter: func(d entity.Domain) string {
			return d.Name + " " + d.Registrar + " " + strings.Join(d.Tags, ", ")
		},
		open: func(d entity.Domain) Screen {
			return newDomainDetail(repo, d, ips)
		},
		add: func() Screen {
			return newDomainForm(repo, entity.Domain{}, ips)
		},
		edit: func(d entity.Domain) Screen {
			return newDomainForm(repo, d, ips)
		},
		remove: func(id int) tea.Cmd {
			return actionCmd(func() error {
				return repo.Delete(id)
			})
		},
		confirm: func(d entity.Domain) string {
			return fmt.Sprintf("Delete %s?", d.Name)
		},
	})
}

func newDomainForm(repo storage.Repo[entity.Domain], original entity.Domain, ips *serverIndex) *formScreen[entity.Domain] {
	return newFormScreen(formConfig[entity.Domain]{
		addTitle:  "Add Domain",
		editTitle: "Edit Domain",
		id: func(d entity.Domain) int {
			return d.ID
		},
		fields: []fieldSpec{
			{label: "Name", placeholder: "example.com"},
			{label: "Server", placeholder: "enter to pick, backspace to clear", pick: func() Screen {
				return newServerPicker()
			}},
			{label: "Registrar"},
			{label: "Expires At", placeholder: "2027-10-01"},
			{label: "Price", placeholder: "in USD"},
			{label: "Notes"},
			{label: "Tags", placeholder: "comma, separated"},
		},
		load: func(d entity.Domain, in []textinput.Model, vals []string) {
			in[0].SetValue(d.Name)
			if d.ServerID != 0 {
				if ips.exists(d.ServerID) {
					vals[serverField] = strconv.Itoa(d.ServerID)
				}
				in[serverField].SetValue(ips.label(d.ServerID))
			}
			in[2].SetValue(d.Registrar)
			in[3].SetValue(d.ExpiresAt)
			in[4].SetValue(pricePtrToString(d.Price))
			in[5].SetValue(d.Notes)
			in[6].SetValue(strings.Join(d.Tags, ", "))
		},
		build: func(id int, in []textinput.Model, vals []string) entity.Domain {
			serverID, _ := strconv.Atoi(vals[serverField])

			price, _ := toPricePtr(in[4].Value())

			return entity.Domain{
				ID:        id,
				Name:      strings.TrimSpace(in[0].Value()),
				ServerID:  serverID,
				Registrar: strings.TrimSpace(in[2].Value()),
				ExpiresAt: strings.TrimSpace(in[3].Value()),
				Price:     price,
				Notes:     strings.TrimSpace(in[5].Value()),
				Tags:      parseTags(in[6].Value()),
			}
		},
		validate: func(d entity.Domain) error {
			return d.Validate()
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
	repo storage.Repo[entity.Domain]
	dom  entity.Domain
	ips  *serverIndex
}

func newDomainDetail(repo storage.Repo[entity.Domain], dom entity.Domain, ips *serverIndex) *domainDetailScreen {
	return &domainDetailScreen{repo: repo, dom: dom, ips: ips}
}

func (d *domainDetailScreen) Title() string {
	return d.dom.Name
}

func (d *domainDetailScreen) Init() tea.Cmd {
	return nil
}

func (d *domainDetailScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return d, nil
	}

	switch {
	case key.Matches(k, keys.Back):
		return d, pop
	case key.Matches(k, keys.Edit):
		return d, push(newDomainForm(d.repo, d.dom, d.ips))
	case key.Matches(k, keys.Renew):
		return d, renewConfirm(d.dom.Name, d.dom.Renewed(time.Now()).ExpiresAt, renewDomainCmd(d.repo, d.dom.ID))
	}

	return d, nil
}

func (d *domainDetailScreen) View() string {
	server := "unassigned"
	if d.dom.ServerID > 0 {
		server = d.ips.label(d.dom.ServerID)
	}
	rows := [][2]string{
		{"Name", d.dom.Name},
		{"Server", server},
		{"Registrar", d.dom.Registrar},
		{"Expires At", d.dom.ExpiresAt},
		{"Price", pricePtrToString(d.dom.Price)},
		{"Notes", d.dom.Notes},
		{"Tags", strings.Join(d.dom.Tags, ", ")},
	}

	days, ok := d.dom.DaysLeft()
	return renderRows(rows, map[string]string{"Expires At": expiryBadge(days, ok)})
}

func (d *domainDetailScreen) ShortHelp() []key.Binding {
	return []key.Binding{keys.Renew, keys.Edit, keys.Back}
}

func (d *domainDetailScreen) FullHelp() [][]key.Binding {
	return [][]key.Binding{{keys.Renew, keys.Edit, keys.Back, keys.Quit}}
}

func domainRowDescription(d entity.Domain, ips *serverIndex) string {
	var parts []string
	if d.Registrar != "" {
		parts = append(parts, d.Registrar)
	}
	if d.ExpiresAt != "" {
		parts = append(parts, d.ExpiresAt)
	}
	if d.ServerID != 0 {
		parts = append(parts, "server "+ips.label(d.ServerID))
	}
	if len(d.Tags) > 0 {
		parts = append(parts, fmt.Sprintf("[%s]", strings.Join(d.Tags, ", ")))
	}

	days, ok := d.DaysLeft()
	return withBadge(strings.Join(parts, " · "), expiryBadge(days, ok))
}

func renewDomainCmd(repo storage.Repo[entity.Domain], id int) tea.Cmd {
	return actionCmd(func() error {
		current, err := repo.Get(id)
		if err != nil {
			return err
		}

		return repo.Update(current.Renewed(time.Now()))
	})
}
