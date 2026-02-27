package plm

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Store struct {
	mu         sync.RWMutex
	productSeq atomic.Uint64
	products   map[string]*Product
	bomSeq     atomic.Uint64
	bomLines   map[string]*BOMLine
	ecoSeq     atomic.Uint64
	ecos       map[string]*ECO
}

func NewStore() *Store {
	return &Store{
		products: make(map[string]*Product),
		bomLines: make(map[string]*BOMLine),
		ecos:     make(map[string]*ECO),
	}
}

func seqID(prefix string, n uint64) string { return prefix + "-" + fmt.Sprintf("%d", n) }

func (s *Store) CreateProduct(p *Product) *Product {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.ID == "" {
		p.ID = seqID("prod", s.productSeq.Add(1))
	}
	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
	if p.Status == "" {
		p.Status = ProductStatusDraft
	}
	cp := *p
	s.products[cp.ID] = &cp
	return &cp
}

func (s *Store) GetProduct(id string) *Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.products[id]
	if !ok {
		return nil
	}
	cp := *p
	return &cp
}

func (s *Store) UpdateProduct(p *Product) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.products[p.ID]; !ok {
		return false
	}
	p.UpdatedAt = time.Now().UTC()
	cp := *p
	s.products[p.ID] = &cp
	return true
}

func (s *Store) ListProducts(status ProductStatus, sku string) []Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Product
	for _, p := range s.products {
		if status != "" && p.Status != status {
			continue
		}
		if sku != "" && p.SKU != sku {
			continue
		}
		out = append(out, *p)
	}
	return out
}

func (s *Store) AddBOMLine(b *BOMLine) *BOMLine {
	s.mu.Lock()
	defer s.mu.Unlock()
	if b.ID == "" {
		b.ID = seqID("bom", s.bomSeq.Add(1))
	}
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now().UTC()
	}
	cp := *b
	s.bomLines[cp.ID] = &cp
	return &cp
}

func (s *Store) GetBOMLine(id string) *BOMLine {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.bomLines[id]
	if !ok {
		return nil
	}
	cp := *b
	return &cp
}

func (s *Store) DeleteBOMLine(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.bomLines[id]; !ok {
		return false
	}
	delete(s.bomLines, id)
	return true
}

func (s *Store) ListBOMLinesByParent(parentProductID string) []BOMLine {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []BOMLine
	for _, b := range s.bomLines {
		if b.ParentProductID != parentProductID {
			continue
		}
		out = append(out, *b)
	}
	return out
}

func (s *Store) CreateECO(e *ECO) *ECO {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e.ID == "" {
		e.ID = seqID("eco", s.ecoSeq.Add(1))
	}
	now := time.Now().UTC()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	e.UpdatedAt = now
	if e.Status == "" {
		e.Status = ECOStatusDraft
	}
	cp := *e
	s.ecos[cp.ID] = &cp
	return &cp
}

func (s *Store) GetECO(id string) *ECO {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.ecos[id]
	if !ok {
		return nil
	}
	cp := *e
	return &cp
}

func (s *Store) UpdateECO(e *ECO) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.ecos[e.ID]; !ok {
		return false
	}
	e.UpdatedAt = time.Now().UTC()
	cp := *e
	s.ecos[e.ID] = &cp
	return true
}

func (s *Store) ListECOs(status ECOStatus, productID string) []ECO {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []ECO
	for _, e := range s.ecos {
		if status != "" && e.Status != status {
			continue
		}
		if productID != "" && e.ProductID != productID {
			continue
		}
		out = append(out, *e)
	}
	return out
}
