package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/skalgutkar/grpc-catalog-platform/services/catalog-service/internal/domain"
)

func seedCatalog() *domain.Catalog {
	return domain.NewCatalog([]domain.Product{
		{ID: "p1", Name: "Widget", UnitPriceCents: 1000},
		{ID: "p2", Name: "Gadget", UnitPriceCents: 2500},
		{ID: "p3", Name: "Gizmo", UnitPriceCents: 500},
	})
}

func TestCatalog_Get(t *testing.T) {
	t.Parallel()

	c := seedCatalog()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		p, err := c.Get("p2")
		require.NoError(t, err)
		assert.Equal(t, "Gadget", p.Name)
		assert.Equal(t, int64(2500), p.UnitPriceCents)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		_, err := c.Get("does-not-exist")
		require.ErrorIs(t, err, domain.ErrProductNotFound)
	})
}

func TestCatalog_List(t *testing.T) {
	t.Parallel()

	c := seedCatalog()

	tests := []struct {
		name     string
		pageSize int32
		wantLen  int
	}{
		{name: "zero means no limit", pageSize: 0, wantLen: 3},
		{name: "negative means no limit", pageSize: -1, wantLen: 3},
		{name: "limit smaller than catalog", pageSize: 2, wantLen: 2},
		{name: "limit larger than catalog", pageSize: 100, wantLen: 3},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := c.List(tt.pageSize)
			assert.Len(t, got, tt.wantLen)
		})
	}
}
