package domain

import "errors"

var (
	ErrInvalidUnitPrice = errors.New("unit price must be positive")
	ErrInvalidQuantity  = errors.New("quantity must be positive")
)

// discountTiers is checked from highest quantity down, so the first
// match wins. Kept as a package-level table rather than nested
// if/else so a new tier is a one-line change, and the table itself is
// what a test asserts against.
var discountTiers = []struct {
	minQuantity int32
	percent     int32
}{
	{minQuantity: 100, percent: 15},
	{minQuantity: 50, percent: 10},
	{minQuantity: 10, percent: 5},
	{minQuantity: 0, percent: 0},
}

// Calculate applies volume-discount pricing. It's pure — no gRPC, no
// context, no I/O — so it's table-tested directly without spinning up a
// server.
func Calculate(unitPriceCents int64, quantity int32) (totalCents int64, discountPercent int32, err error) {
	if unitPriceCents <= 0 {
		return 0, 0, ErrInvalidUnitPrice
	}
	if quantity <= 0 {
		return 0, 0, ErrInvalidQuantity
	}

	discountPercent = discountForQuantity(quantity)

	subtotal := unitPriceCents * int64(quantity)
	discount := subtotal * int64(discountPercent) / 100
	return subtotal - discount, discountPercent, nil
}

func discountForQuantity(quantity int32) int32 {
	for _, tier := range discountTiers {
		if quantity >= tier.minQuantity {
			return tier.percent
		}
	}
	return 0
}
