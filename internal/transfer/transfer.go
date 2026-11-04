package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/omidnikrah/rack-cli/internal/entity"
	"github.com/omidnikrah/rack-cli/internal/server"
	"github.com/omidnikrah/rack-cli/internal/storage"
)

const formatVersion = 1

var dataFiles = []string{"servers.json", "domains.json"}

type serverRecord struct {
	entity.Server
	Password string `json:"password,omitempty"`
}

type bundle struct {
	Version    int             `json:"version"`
	ExportedAt time.Time       `json:"exported_at"`
	Servers    []serverRecord  `json:"servers"`
	Domains    []entity.Domain `json:"domains"`
}

type Result struct {
	Servers int
	Domains int
	Skipped int
}

func Export(w io.Writer, servers *server.Service, domains storage.Repo[entity.Domain], withPasswords bool) error {
	srvs, err := servers.Load()
	if err != nil {
		return err
	}

	doms, err := domains.Load()
	if err != nil {
		return err
	}

	b := bundle{
		Version:    formatVersion,
		ExportedAt: time.Now().UTC(),
		Servers:    make([]serverRecord, 0, len(srvs)),
		Domains:    append([]entity.Domain{}, doms...),
	}
	for _, s := range srvs {
		rec := serverRecord{Server: s}
		if withPasswords {
			if rec.Password, err = servers.Password(s); err != nil {
				return fmt.Errorf("reading password of %s: %w", s.Label, err)
			}
		}
		b.Servers = append(b.Servers, rec)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(b)
}

func Import(r io.Reader, servers *server.Service, domains storage.Repo[entity.Domain]) (Result, error) {
	var b bundle
	if err := json.NewDecoder(r).Decode(&b); err != nil {
		return Result{}, fmt.Errorf("reading export: %w", err)
	}

	if b.Version != formatVersion {
		return Result{}, fmt.Errorf("unsupported export version %d", b.Version)
	}

	if err := validate(b); err != nil {
		return Result{}, err
	}

	existingServers, err := servers.Load()
	if err != nil {
		return Result{}, err
	}

	existingDomains, err := domains.Load()
	if err != nil {
		return Result{}, err
	}

	var res Result
	serverIDs := make(map[int]int)
	byLabel := make(map[string]int, len(existingServers))
	for _, s := range existingServers {
		byLabel[key(s.Label)] = s.ID
	}

	for _, rec := range b.Servers {
		if id, ok := byLabel[key(rec.Label)]; ok {
			serverIDs[rec.ID] = id
			res.Skipped++
			continue
		}

		srv := rec.Server
		srv.ID, srv.Password = 0, ""
		saved, err := servers.Save(srv, rec.Password)
		if err != nil {
			return res, fmt.Errorf("importing server %s: %w", rec.Label, err)
		}

		serverIDs[rec.ID] = saved.ID
		byLabel[key(saved.Label)] = saved.ID
		res.Servers++
	}

	names := make(map[string]bool, len(existingDomains))
	for _, d := range existingDomains {
		names[key(d.Name)] = true
	}

	for _, d := range b.Domains {
		if names[key(d.Name)] {
			res.Skipped++
			continue
		}

		d.ID, d.ServerID = 0, serverIDs[d.ServerID]
		if _, err := storage.Save(domains, d); err != nil {
			return res, fmt.Errorf("importing domain %s: %w", d.Name, err)
		}

		names[key(d.Name)] = true
		res.Domains++
	}

	return res, nil
}

func Backup(dataDir, backupRoot string) (string, error) {
	dest := filepath.Join(backupRoot, "rack-cli-"+time.Now().Format("20060102-150405"))
	if err := os.MkdirAll(dest, 0o700); err != nil {
		return "", err
	}

	copied := 0
	for _, name := range dataFiles {
		raw, err := os.ReadFile(filepath.Join(dataDir, name))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", err
		}

		if err := os.WriteFile(filepath.Join(dest, name), raw, 0o600); err != nil {
			return "", err
		}
		copied++
	}

	if copied == 0 {
		os.Remove(dest)
		return "", errors.New("nothing to back up yet")
	}

	return dest, nil
}

func validate(b bundle) error {
	for _, rec := range b.Servers {
		srv := rec.Server
		srv.ID = 0
		if err := srv.Validate(); err != nil {
			return fmt.Errorf("server %q: %w", rec.Label, err)
		}
	}

	for _, d := range b.Domains {
		if err := d.Validate(); err != nil {
			return fmt.Errorf("domain %q: %w", d.Name, err)
		}
	}

	return nil
}

func key(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
