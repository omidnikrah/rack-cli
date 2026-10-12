package entity

import (
	"errors"
	"strings"
	"time"
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
	return validateDate(d.ExpiresAt, "expires at")
}

func (d Domain) DaysLeft() (days int, ok bool) {
	return d.DaysLeftAt(time.Now())
}

func (d Domain) DaysLeftAt(now time.Time) (days int, ok bool) {
	return daysUntil(d.ExpiresAt, now)
}
