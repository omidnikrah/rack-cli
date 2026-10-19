package tui

import (
	"fmt"
	"math"
	"strconv"
)

func toPricePtr(p string) (*uint, error) {
	if p == "" {
		return nil, nil
	}
	price, err := strconv.ParseFloat(p, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid price format (use format like: 12.50)")
	}
	if price < 0 {
		return nil, fmt.Errorf("price cannot be negative")
	}
	priceVal := uint(math.Round(price * 100))
	return &priceVal, nil
}

func pricePtrToString(p *uint) string {
	if p == nil {
		return ""
	}
	price := float64(*p) / 100
	return fmt.Sprintf("%.2f", price)
}
