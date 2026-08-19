package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/skalgutkar/grpc-catalog-platform/services/pricing-service/internal/domain"
)

func TestCalculate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		unitPriceCents int64
		quantity       int32
		wantTotalCents int64
		wantDiscount   int32
		wantErr        error
	}{
		{
			name:           "below first tier gets no discount",
			unitPriceCents: 1000,
			quantity:       9,
			wantTotalCents: 9000,
			wantDiscount:   0,
		},
		{
			name:           "exactly at 10 gets 5 percent",
			unitPriceCents: 1000,
			quantity:       10,
			wantTotalCents: 9500,
			wantDiscount:   5,
		},
		{
			name:           "exactly at 50 gets 10 percent",
			unitPriceCents: 1000,
			quantity:       50,
			wantTotalCents: 45000,
			wantDiscount:   10,
		},
		{
			name:           "exactly at 100 gets 15 percent",
			unitPriceCents: 1000,
			quantity:       100,
			wantTotalCents: 85000,
			wantDiscount:   15,
		},
		{
			name:           "well above top tier stays at 15 percent",
			unitPriceCents: 1000,
			quantity:       1000,
			wantTotalCents: 850000,
			wantDiscount:   15,
		},
		{
			name:           "zero unit price is rejected",
			unitPriceCents: 0,
			quantity:       1,
			wantErr:        domain.ErrInvalidUnitPrice,
		},
		{
			name:           "negative unit price is rejected",
			unitPriceCents: -100,
			quantity:       1,
			wantErr:        domain.ErrInvalidUnitPrice,
		},
		{
			name:           "zero quantity is rejected",
			unitPriceCents: 1000,
			quantity:       0,
			wantErr:        domain.ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			total, discount, err := domain.Calculate(tt.unitPriceCents, tt.quantity)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantTotalCents, total)
			assert.Equal(t, tt.wantDiscount, discount)
		})
	}
}
