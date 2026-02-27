package plm

import (
	"context"
	"fmt"
)

// Service implements PLM business logic.
type Service struct {
	Store *Store
}

// NewService returns a PLM service.
func NewService(store *Store) *Service {
	return &Service{Store: store}
}

// CreateProduct creates a product.
func (s *Service) CreateProduct(ctx context.Context, p *Product) (*Product, error) {
	if p.SKU == "" {
		return nil, fmt.Errorf("sku required")
	}
	if p.Name == "" {
		return nil, fmt.Errorf("name required")
	}
	if p.Revision == "" {
		p.Revision = "A"
	}
	return s.Store.CreateProduct(p), nil
}

// GetProduct returns a product by ID.
func (s *Service) GetProduct(ctx context.Context, id string) *Product {
	return s.Store.GetProduct(id)
}

// UpdateProduct updates a product.
func (s *Service) UpdateProduct(ctx context.Context, p *Product) (*Product, error) {
	if s.Store.GetProduct(p.ID) == nil {
		return nil, fmt.Errorf("product not found: %s", p.ID)
	}
	s.Store.UpdateProduct(p)
	return s.Store.GetProduct(p.ID), nil
}

// ListProducts returns products with optional filters.
func (s *Service) ListProducts(ctx context.Context, status ProductStatus, sku string) []Product {
	return s.Store.ListProducts(status, sku)
}

// AddBOMLine adds a BOM line to a product.
func (s *Service) AddBOMLine(ctx context.Context, parentProductID string, b *BOMLine) (*BOMLine, error) {
	if s.Store.GetProduct(parentProductID) == nil {
		return nil, fmt.Errorf("product not found: %s", parentProductID)
	}
	if b.ChildSKU == "" {
		return nil, fmt.Errorf("child_sku required")
	}
	if b.Quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	b.ParentProductID = parentProductID
	return s.Store.AddBOMLine(b), nil
}

// DeleteBOMLine removes a BOM line.
func (s *Service) DeleteBOMLine(ctx context.Context, lineID string) error {
	if !s.Store.DeleteBOMLine(lineID) {
		return fmt.Errorf("bom line not found: %s", lineID)
	}
	return nil
}

// GetBOM returns all BOM lines for a product.
func (s *Service) GetBOM(ctx context.Context, parentProductID string) []BOMLine {
	return s.Store.ListBOMLinesByParent(parentProductID)
}

// CreateECO creates an engineering change order.
func (s *Service) CreateECO(ctx context.Context, e *ECO) (*ECO, error) {
	if e.Title == "" {
		return nil, fmt.Errorf("title required")
	}
	return s.Store.CreateECO(e), nil
}

// GetECO returns an ECO by ID.
func (s *Service) GetECO(ctx context.Context, id string) *ECO {
	return s.Store.GetECO(id)
}

// UpdateECO updates an ECO.
func (s *Service) UpdateECO(ctx context.Context, e *ECO) (*ECO, error) {
	if s.Store.GetECO(e.ID) == nil {
		return nil, fmt.Errorf("eco not found: %s", e.ID)
	}
	s.Store.UpdateECO(e)
	return s.Store.GetECO(e.ID), nil
}

// ListECOs returns ECOs with optional filters.
func (s *Service) ListECOs(ctx context.Context, status ECOStatus, productID string) []ECO {
	return s.Store.ListECOs(status, productID)
}
