package entity

import (
	"errors"
	"strings"
)

type Server struct {
	ID       int      `json:"id"`
	Label    string   `json:"label"`
	IP       string   `json:"ip"`
	RootUser string   `json:"root_user"`
	Password string   `json:"password"`
	Purpose  string   `json:"purpose"`
	Tags     []string `json:"tags"`
}

func (s Server) GetID() int {
	return s.ID
}

func (s Server) SetID(id int) Server {
	s.ID = id
	return s
}

func (s Server) Validate() error {
	if strings.TrimSpace(s.Label) == "" {
		return errors.New("label is required")
	}

	if strings.TrimSpace(s.IP) == "" {
		return errors.New("IP address is required")
	}

	if strings.TrimSpace(s.RootUser) == "" {
		return errors.New("root user is required")
	}

	return nil
}
