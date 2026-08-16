package menu

import (
	"rack-cli/internal/domain"
)

func runDomains() []MenuItems {
	return []MenuItems{
		{Label: "Add new domain", Action: domain.Add},
		{Label: "Edit a domain", Action: domain.Edit},
		{Label: "Delete a domain", Action: domain.Delete},
		{Label: "View a domain", Action: domain.View},
		{Label: "Ping a domain", Action: domain.Ping},
		{Label: "List domains", Action: domain.List},
	}
}
