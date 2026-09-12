package main

import (
	"fmt"
	"os"
	"path/filepath"
	"rack-cli/internal/entity"
	"rack-cli/internal/storage"
	"rack-cli/internal/tui"

	tea "charm.land/bubbletea/v2"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "rack-cli: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	dir, err := storage.DefaultDir()
	if err != nil {
		return fmt.Errorf("locating config dir: %w", err)
	}

	servers := storage.NewJSONRepo[entity.Server](filepath.Join(dir, "servers.json"))
	domains := storage.NewJSONRepo[entity.Domain](filepath.Join(dir, "domains.json"))

	_, err = tea.NewProgram(tui.New(servers, domains)).Run()
	return err
}
