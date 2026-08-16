package menu

import (
	"rack-cli/internal/server"
	"rack-cli/internal/terminal"
)

func Run(storage server.Storage, tml terminal.Terminal) {
	RenderMenu("** Welcome to Rack CLI **", []MenuItems{
		{Label: "Manage Servers", SubTitle: "** Manage Servers **", SubItems: func() []MenuItems { return runServers(storage, tml) }},
		{Label: "Manage Domains", SubTitle: "** Manage Domains **", SubItems: func() []MenuItems { return runDomains() }},
	}, tml)
}
