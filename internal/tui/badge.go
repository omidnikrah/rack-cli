package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/omidnikrah/rack-cli/internal/expiry"
)

const soonDays = 30

var (
	dangerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#E5675B")).Bold(true)
	warnStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#E9A23B"))
)

func expiryBadge(days int, ok bool) string {
	if !ok {
		return ""
	}

	switch {
	case days < 0:
		return dangerStyle.Render(fmt.Sprintf("expired %dd ago", -days))
	case days == 0:
		return dangerStyle.Render("expires today")
	case expiry.IsDue(days):
		return dangerStyle.Render(daysIn(days))
	case days <= soonDays:
		return warnStyle.Render(daysIn(days))
	default:
		return okStyle.Render(daysIn(days))
	}
}

func daysIn(days int) string {
	if days == 1 {
		return "tomorrow"
	}
	if days >= 100 {
		return fmt.Sprintf("in %dmo", days/30)
	}
	return fmt.Sprintf("in %dd", days)
}

func withBadge(desc, badge string) string {
	if badge == "" {
		return desc
	}
	return desc + " · " + badge
}

func renderRows(rows [][2]string, badges map[string]string) string {
	var b strings.Builder
	for _, row := range rows {
		b.WriteString(labelStyle.Render(row[0]))
		b.WriteString(valueStyle.Render(row[1]))
		if badge := badges[row[0]]; badge != "" {
			b.WriteString("  " + badge)
		}
		b.WriteString("\n")
	}
	return b.String()
}
