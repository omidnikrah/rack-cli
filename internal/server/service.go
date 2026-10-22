package server

import (
	"github.com/omidnikrah/rack-cli/internal/entity"
	"github.com/omidnikrah/rack-cli/internal/secret"
)

type Service struct {
	Storage
	vault *secret.Vault
}

func NewService(repo Storage, vault *secret.Vault) *Service {
	return &Service{Storage: repo, vault: vault}
}

func (s *Service) Save(srv entity.Server, password string) (entity.Server, error) {
	if err := srv.Validate(); err != nil {
		return srv, err
	}

	if srv.ID == 0 {
		srv.Password = ""
		created, err := s.Create(srv)
		if err != nil {
			return srv, err
		}

		srv = created
	} else if password == "" {
		current, err := s.Get(srv.ID)
		if err != nil {
			return srv, err
		}
		srv.Password = current.Password
		return srv, s.Update(srv)
	}

	sealed, err := s.vault.Seal(password, srv.ID)
	if err != nil {
		return srv, err
	}

	srv.Password = sealed
	return srv, s.Update(srv)
}

func (s *Service) Password(srv entity.Server) (string, error) {
	return s.vault.Open(srv.Password, srv.ID)
}
