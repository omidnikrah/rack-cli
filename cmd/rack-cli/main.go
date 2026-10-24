package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/omidnikrah/rack-cli/internal/entity"
	"github.com/omidnikrah/rack-cli/internal/expiry"
	"github.com/omidnikrah/rack-cli/internal/notify"
	"github.com/omidnikrah/rack-cli/internal/schedule"
	"github.com/omidnikrah/rack-cli/internal/secret"
	"github.com/omidnikrah/rack-cli/internal/server"
	"github.com/omidnikrah/rack-cli/internal/storage"
	"github.com/omidnikrah/rack-cli/internal/tui"
	"io"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
)

// version is set at build time by goreleaser.
var version = "dev"

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

	check := flag.Bool("check", false, "print servers close to expiry and send notifications")
	enable := flag.Bool("enable-notify", false, "run the expiry check daily at 09:00 (macOS)")
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("rack-cli", version)
		return nil
	}

	serverRepo := storage.NewJSONRepo[entity.Server](filepath.Join(dir, "servers.json"))
	domainRepo := storage.NewJSONRepo[entity.Domain](filepath.Join(dir, "domains.json"))

	switch {
	case *check:
		return runExpiry(os.Stdout, serverRepo, domainRepo)
	case *enable:
		return enableNotify(os.Stdout, dir)
	}

	vault := secret.NewVault(secret.Keychain{})
	servers := server.NewService(serverRepo, vault)
	_, err = tea.NewProgram(tui.New(servers, domainRepo)).Run()
	return err
}

func runExpiry(w io.Writer, serverRepo storage.Repo[entity.Server], domainRepo storage.Repo[entity.Domain]) error {
	servers, err := serverRepo.Load()
	if err != nil {
		return err
	}

	domains, err := domainRepo.Load()
	if err != nil {
		return err
	}

	msgs := expiry.Messages(servers, domains)
	if len(msgs) == 0 {
		fmt.Fprintln(w, "No servers expiring soon.")
	}

	var errs []error
	for _, msg := range msgs {
		fmt.Fprintln(w, msg)
		if err := notify.Send("Server expiring", msg); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func enableNotify(w io.Writer, dir string) error {
	if err := schedule.EnsureInstalled(dir); err != nil {
		return err
	}

	fmt.Fprintln(w, "Expiry notifications enabled: checking daily at 09:00.")
	return nil
}
