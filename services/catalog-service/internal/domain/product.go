package domain

import "errors"

var ErrProductNotFound = errors.New("product not found")

type Product struct {
	ID             string
	Name           string
	UnitPriceCents int64
}

// Catalog is an in-memory product store. A real system would back this
// with Postgres/DynamoDB; here the interesting thing under test is the
// gRPC/gateway/pricing-client wiring around it, not a second CRUD layer —
// same tradeoff as inventory-service's FixedStock in the sibling repo.
type Catalog struct {
	products map[string]Product
	order    []string // preserves insertion order for ListProducts
}

func NewCatalog(seed []Product) *Catalog {
	c := &Catalog{products: make(map[string]Product, len(seed))}
	for _, p := range seed {
		c.products[p.ID] = p
		c.order = append(c.order, p.ID)
	}
	return c
}

func (c *Catalog) Get(id string) (Product, error) {
	p, ok := c.products[id]
	if !ok {
		return Product{}, ErrProductNotFound
	}
	return p, nil
}

// List returns up to pageSize products in insertion order. pageSize <= 0
// means "no explicit limit" and returns everything.
func (c *Catalog) List(pageSize int32) []Product {
	limit := len(c.order)
	if pageSize > 0 && int(pageSize) < limit {
		limit = int(pageSize)
	}

	result := make([]Product, 0, limit)
	for _, id := range c.order[:limit] {
		result = append(result, c.products[id])
	}
	return result
}
