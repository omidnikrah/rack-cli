package menu

import (
	"fmt"
	"rack-cli/internal/server"
)

func runServers(storage server.Storage, in server.InputReader) []MenuItems {
	return []MenuItems{
		{Label: "Add new server", Action: func() {
			server.Add(storage, in)
		}},
		{Label: "Edit a server", SubTitle: "** Enter server number to edit **", SubItems: func() []MenuItems {
			return renderServersList(storage)
		}},
		{Label: "Delete a server", Action: server.Delete},
		{Label: "View a server", Action: server.View},
		{Label: "Ping a server", Action: server.Ping},
		{Label: "SSH a server", Action: server.SSH},
		{Label: "List servers", SubItems: func() []MenuItems {
			return renderServersList(storage)
		}},
	}
}

func renderServersList(storage server.Storage) []MenuItems {
	servers, err := storage.Load()
	if err != nil {
		fmt.Println(err)
		return []MenuItems{}
	}

	menuItems := make([]MenuItems, 0, len(servers))
	for _, s := range servers {
		menuItems = append(menuItems, MenuItems{Label: s.Label, Action: func() { server.Edit(storage, s) }})
	}

	return menuItems
}
