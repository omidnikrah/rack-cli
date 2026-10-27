package tui

import (
	"fmt"
	"math"
	"strconv"

	"github.com/omidnikrah/rack-cli/internal/entity"
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
	priceVal := uint(min(math.Round(price*100), float64(entity.MaxPrice)+1))
	return &priceVal, nil
}

func pricePtrToString(p *uint) string {
	if p == nil {
		return ""
	}
	price := float64(*p) / 100
	return fmt.Sprintf("%.2f", price)
}
