package tui

import (
	"github.com/omidnikrah/rack-cli/internal/entity"
	"github.com/omidnikrah/rack-cli/internal/expiry"
	"strconv"

	"charm.land/lipgloss/v2"
)

type overviewStats struct {
	servers      int
	domains      int
	expiringSoon int
	totalPrice   uint
}

func computeStats(servers []entity.Server, domains []entity.Domain) overviewStats {
	st := overviewStats{servers: len(servers), domains: len(domains)}

	for _, s := range servers {
		if days, ok := s.DaysLeft(); ok && expiry.IsDue(days) {
			st.expiringSoon++
		}
		if s.Price != nil {
			st.totalPrice += *s.Price
		}
	}

	for _, d := range domains {
		if days, ok := d.DaysLeft(); ok && expiry.IsDue(days) {
			st.expiringSoon++
		}
		if d.Price != nil {
			st.totalPrice += *d.Price
		}
	}

	return st
}

var overviewHeight = lipgloss.Height(renderOverview(overviewStats{}))

func tile(label, value string, accent lipgloss.Style) string {
	return tileStyle.Render(lipgloss.JoinVertical(lipgloss.Left, blurredStyle.Render(label), accent.Bold(true).Render(value)))
}

func renderOverview(st overviewStats) string {
	expiring := okStyle

	if st.expiringSoon > 0 {
		expiring = dangerStyle
	}

	return lipgloss.JoinHorizontal(lipgloss.Top,
		tile("Servers", strconv.Itoa(st.servers), focusedStyle),
		tile("Domains", strconv.Itoa(st.domains), focusedStyle),
		tile("Expiring", strconv.Itoa(st.expiringSoon), expiring),
		tile("Total price", "$"+pricePtrToString(&st.totalPrice), valueStyle),
	)
}
