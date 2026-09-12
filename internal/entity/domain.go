package entity

import (
	"errors"
	"strings"
)

type Domain struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	ServerID  int      `json:"server_id"`
	Registrar string   `json:"registrar"`
	ExpiresAt string   `json:"expires_at"`
	Notes     string   `json:"notes"`
	Tags      []string `json:"tags"`
}

func (d Domain) GetID() int { return d.ID }

func (d Domain) SetID(id int) Domain {
	d.ID = id
	return d
}

func (d Domain) Validate() error {
	if strings.TrimSpace(d.Name) == "" {
		return errors.New("name is required")
	}
	if strings.Contains(d.Name, " ") {
		return errors.New("name cannot contain spaces")
	}
	return nil
}
