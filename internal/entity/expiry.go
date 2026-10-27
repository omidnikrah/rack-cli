package entity

import (
	"errors"
	"math"
	"strings"
	"time"
)

const DateLayout = "2006-01-02"

var errNoDate = errors.New("no date set")

func parseDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, errNoDate
	}

	return time.ParseInLocation(DateLayout, raw, time.Local)
}

func validateDate(raw, field string) error {
	if _, err := parseDate(raw); err != nil && !errors.Is(err, errNoDate) {
		return errors.New(field + " must look like 2025-12-31")
	}

	return nil
}

func daysUntil(raw string, now time.Time) (days int, ok bool) {
	date, err := parseDate(raw)
	if err != nil {
		return 0, false
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	return int(math.Round(date.Sub(today).Hours() / 24)), true
}

func renewDate(raw string, now time.Time, months int) string {
	base := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	if date, err := parseDate(raw); err == nil && date.After(base) {
		base = date
	}

	return addMonths(base, months).Format(DateLayout)
}

func addMonths(t time.Time, months int) time.Time {
	first := time.Date(t.Year(), t.Month()+time.Month(months), 1, 0, 0, 0, 0, time.Local)
	lastDay := first.AddDate(0, 1, -1).Day()

	return time.Date(first.Year(), first.Month(), min(t.Day(), lastDay), 0, 0, 0, 0, time.Local)
}
