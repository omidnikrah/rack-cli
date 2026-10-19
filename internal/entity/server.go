package entity

import (
	"errors"
	"net"
	"rack-cli/internal/secret"
	"strings"
	"time"
)

type Server struct {
	ID        int           `json:"id"`
	Label     string        `json:"label"`
	IP        string        `json:"ip"`
	RootUser  string        `json:"root_user"`
	Password  secret.Sealed `json:"password"`
	Purpose   string        `json:"purpose"`
	Price     *uint         `json:"price"`
	ExpiredAt string        `json:"expired_at"`
	Tags      []string      `json:"tags"`
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

	if net.ParseIP(s.IP) == nil {
		return errors.New("invalid IP address")
	}

	if strings.TrimSpace(s.RootUser) == "" {
		return errors.New("root user is required")
	}

	if s.Price != nil && *s.Price == 0 {
		return errors.New("price must be greater than 0")
	}

	return validateDate(s.ExpiredAt, "expired at")
}

func (s Server) DaysLeft() (days int, ok bool) {
	return s.DaysLeftAt(time.Now())
}

func (s Server) DaysLeftAt(now time.Time) (days int, ok bool) {
	return daysUntil(s.ExpiredAt, now)
}
