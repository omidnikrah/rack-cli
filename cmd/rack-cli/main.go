package main

import (
	"os"

	"rack-cli/internal/entity"
	"rack-cli/internal/menu"
	"rack-cli/internal/storage"
	"rack-cli/internal/terminal"
)

func main() {
	sStorage := storage.New[entity.Server](".")
	tml := terminal.New(os.Stdin)
	menu.Run(sStorage, tml)
}
