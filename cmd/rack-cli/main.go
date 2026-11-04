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
	"github.com/omidnikrah/rack-cli/internal/transfer"
	"github.com/omidnikrah/rack-cli/internal/tui"
	"github.com/omidnikrah/rack-cli/internal/upgrade"
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
	doUpgrade := flag.Bool("upgrade", false, "install the latest release over this binary")
	exportPath := flag.String("export", "", "write servers and domains to a JSON file (- for stdout)")
	importPath := flag.String("import", "", "add servers and domains from a JSON export, skipping ones that already exist")
	withPasswords := flag.Bool("with-passwords", false, "include passwords, as plain text, in --export")
	backupDir := flag.String("backup", "", "copy the data files into a new timestamped folder inside this directory")
	flag.Parse()

	if *showVersion {
		fmt.Println("rack-cli", version)
		return nil
	}

	if *doUpgrade {
		return runUpgrade(os.Stdout)
	}

	serverRepo := storage.NewJSONRepo[entity.Server](filepath.Join(dir, "servers.json"))
	domainRepo := storage.NewJSONRepo[entity.Domain](filepath.Join(dir, "domains.json"))

	vault := secret.NewVault(secret.Keychain{})
	servers := server.NewService(serverRepo, vault)

	switch {
	case *check:
		return runExpiry(os.Stdout, serverRepo, domainRepo)
	case *enable:
		return enableNotify(os.Stdout, dir)
	case *exportPath != "":
		return runExport(os.Stdout, *exportPath, *withPasswords, servers, domainRepo)
	case *importPath != "":
		return runImport(os.Stdout, *importPath, servers, domainRepo)
	case *backupDir != "":
		return runBackup(os.Stdout, dir, *backupDir)
	}

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

func runExport(w io.Writer, path string, withPasswords bool, servers *server.Service, domainRepo storage.Repo[entity.Domain]) error {
	if path == "-" {
		return transfer.Export(w, servers, domainRepo, withPasswords)
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}

	if err := transfer.Export(f, servers, domainRepo, withPasswords); err != nil {
		f.Close()
		os.Remove(path)
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	fmt.Fprintf(w, "Exported to %s\n", path)
	return nil
}

func runImport(w io.Writer, path string, servers *server.Service, domainRepo storage.Repo[entity.Domain]) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	res, err := transfer.Import(f, servers, domainRepo)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Imported %d servers and %d domains, skipped %d that already exist.\n", res.Servers, res.Domains, res.Skipped)
	return nil
}

func runBackup(w io.Writer, dataDir, backupRoot string) error {
	dest, err := transfer.Backup(dataDir, backupRoot)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Backed up to %s\n", dest)
	return nil
}

func runUpgrade(w io.Writer) error {
	res, err := upgrade.Run(version)
	if err != nil {
		return err
	}

	if !res.Upgraded {
		fmt.Fprintf(w, "Already on the latest version (%s).\n", res.To)
		return nil
	}

	fmt.Fprintf(w, "Upgraded %s from %s to %s.\n", res.Path, res.From, res.To)
	return nil
}
