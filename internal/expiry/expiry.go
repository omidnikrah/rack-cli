package expiry

import (
	"fmt"
	"rack-cli/internal/entity"
)

const (
	WarnDays      = 7
	maxIndividual = 3
)

func IsDue(days int) bool {
	return days <= WarnDays
}

func Messages(servers []entity.Server, domains []entity.Domain) []string {
	var msgs []string

	for _, s := range servers {
		if days, ok := s.DaysLeft(); ok && IsDue(days) {
			msgs = append(msgs, describe("Server", s.Label, days))
		}
	}

	for _, d := range domains {
		if days, ok := d.DaysLeft(); ok && IsDue(days) {
			msgs = append(msgs, describe("Domain", d.Name, days))
		}
	}

	if len(msgs) > maxIndividual {
		return []string{fmt.Sprintf("%d servers and domains are expired or expiring within %d days", len(msgs), WarnDays)}
	}

	return msgs
}

func describe(kind, name string, days int) string {
	switch {
	case days < 0:
		return fmt.Sprintf("%s %s expired %d day(s) ago", kind, name, -days)
	case days == 0:
		return fmt.Sprintf("%s %s expires today", kind, name)
	case days == 1:
		return fmt.Sprintf("%s %s expires tomorrow", kind, name)
	default:
		return fmt.Sprintf("%s %s expires in %d days", kind, name, days)
	}
}
