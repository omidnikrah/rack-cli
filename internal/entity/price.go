package entity

import (
	"errors"
	"math"
)

const MaxPrice = math.MaxUint32

func validatePrice(price *uint) error {
	switch {
	case price == nil:
		return nil
	case *price == 0:
		return errors.New("price must be greater than 0")
	case *price > MaxPrice:
		return errors.New("price is too large")
	}

	return nil
}
