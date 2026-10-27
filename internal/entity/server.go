package entity

import (
	"errors"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/omidnikrah/rack-cli/internal/secret"
)

type ServerType string

const (
	TypeVPS       ServerType = "vps"
	TypeShared    ServerType = "shared"
	TypeDedicated ServerType = "dedicated"
	TypeOther     ServerType = "other"
)

func ServerTypes() []ServerType {
	return []ServerType{TypeVPS, TypeShared, TypeDedicated, TypeOther}
}

func (t ServerType) Normalize() ServerType {
	if t == "" {
		return TypeVPS
	}
	return t
}

func (t ServerType) Valid() bool {
	return slices.Contains(ServerTypes(), t.Normalize())
}

func (t ServerType) Label() string {
	switch t.Normalize() {
	case TypeVPS:
		return "VPS"
	case TypeShared:
		return "Shared host"
	case TypeDedicated:
		return "Dedicated"
	default:
		return "Other"
	}
}

func (t ServerType) Description() string {
	switch t.Normalize() {
	case TypeVPS:
		return "Virtual server with root access"
	case TypeShared:
		return "Panel-managed hosting, no root access"
	case TypeDedicated:
		return "A whole physical machine"
	default:
		return "Anything else"
	}
}

func (t ServerType) HasRootAccess() bool {
	t = t.Normalize()
	return t == TypeVPS || t == TypeDedicated
}

type BillingCycle string

const (
	BillingMonthly   BillingCycle = "monthly"
	BillingQuarterly BillingCycle = "quarterly"
	BillingYearly    BillingCycle = "yearly"
	BillingOneTime   BillingCycle = "one-time"
)

func BillingCycles() []BillingCycle {
	return []BillingCycle{BillingMonthly, BillingQuarterly, BillingYearly, BillingOneTime}
}

func (c BillingCycle) Normalize() BillingCycle {
	if c == "" {
		return BillingMonthly
	}
	return c
}

func (c BillingCycle) Valid() bool {
	return slices.Contains(BillingCycles(), c.Normalize())
}

func (c BillingCycle) Label() string {
	switch c.Normalize() {
	case BillingMonthly:
		return "Monthly"
	case BillingQuarterly:
		return "Quarterly"
	case BillingYearly:
		return "Yearly"
	default:
		return "One-time"
	}
}

func (c BillingCycle) Description() string {
	switch c.Normalize() {
	case BillingMonthly:
		return "Charged every month, counted 12x a year"
	case BillingQuarterly:
		return "Charged every 3 months, counted 4x a year"
	case BillingYearly:
		return "Charged once a year"
	default:
		return "Paid once, left out of the yearly cost"
	}
}

func (c BillingCycle) PerLabel() string {
	switch c.Normalize() {
	case BillingMonthly:
		return "/ month"
	case BillingQuarterly:
		return "/ quarter"
	case BillingYearly:
		return "/ year"
	default:
		return "one-time"
	}
}

func (c BillingCycle) timesPerYear() uint {
	if months := c.months(); months > 0 {
		return uint(12 / months)
	}

	return 0
}

func (c BillingCycle) months() int {
	switch c.Normalize() {
	case BillingMonthly:
		return 1
	case BillingQuarterly:
		return 3
	case BillingYearly:
		return 12
	}

	return 0
}

type Server struct {
	ID        int           `json:"id"`
	Label     string        `json:"label"`
	Type      ServerType    `json:"type"`
	IP        string        `json:"ip"`
	RootUser  string        `json:"root_user"`
	Password  secret.Sealed `json:"password"`
	Provider  string        `json:"provider"`
	Purpose   string        `json:"purpose"`
	Billing   BillingCycle  `json:"billing"`
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

	if !s.Type.Valid() {
		return errors.New("invalid server type")
	}

	if !s.Billing.Valid() {
		return errors.New("invalid billing cycle")
	}

	if s.Type.HasRootAccess() {
		if strings.TrimSpace(s.IP) == "" {
			return errors.New("IP address is required")
		}

		if strings.TrimSpace(s.RootUser) == "" {
			return errors.New("root user is required")
		}
	}

	if ip := strings.TrimSpace(s.IP); ip != "" && net.ParseIP(ip) == nil {
		return errors.New("invalid IP address")
	}

	if strings.HasPrefix(strings.TrimSpace(s.RootUser), "-") {
		return errors.New("root user cannot start with a dash")
	}

	if err := validatePrice(s.Price); err != nil {
		return err
	}

	return validateDate(s.ExpiredAt, "expired at")
}

func (s Server) HasHost() bool {
	return strings.TrimSpace(s.IP) != ""
}

func (s Server) YearlyPrice() uint {
	if s.Price == nil {
		return 0
	}
	return *s.Price * s.Billing.timesPerYear()
}

func (s Server) DaysLeft() (days int, ok bool) {
	return s.DaysLeftAt(time.Now())
}

func (s Server) DaysLeftAt(now time.Time) (days int, ok bool) {
	return daysUntil(s.ExpiredAt, now)
}

func (s Server) CanRenew() bool {
	return s.Billing.months() > 0
}

func (s Server) Renewed(now time.Time) Server {
	s.ExpiredAt = renewDate(s.ExpiredAt, now, s.Billing.months())
	return s
}
