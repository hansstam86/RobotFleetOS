package mes

import (
	"sync"
	"sync/atomic"
	"time"
)

// Store is an in-memory store for production orders.
type Store struct {
	mu     sync.RWMutex
	orders map[string]*ProductionOrder
	seq    atomic.Uint64
}

// NewStore returns a new in-memory order store.
func NewStore() *Store {
	return &Store{orders: make(map[string]*ProductionOrder)}
}

// Create persists a new order (ID generated if empty). Returns the order with ID set.
func (s *Store) Create(order *ProductionOrder) (*ProductionOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if order.ID == "" {
		order.ID = "ord-" + formatSeq(s.seq.Add(1))
	}
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now().UTC()
	}
	if order.Status == "" {
		order.Status = OrderStatusDraft
	}
	order.QuantityCompleted = 0
	order.QuantityScrapped = 0
	cp := *order
	s.orders[order.ID] = &cp
	return &cp, nil
}

func formatSeq(n uint64) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// Get returns an order by ID, or nil if not found.
func (s *Store) Get(id string) *ProductionOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	if !ok {
		return nil
	}
	cp := *o
	return &cp
}

// List returns all orders, optionally filtered by status.
func (s *Store) List(statusFilter OrderStatus) []ProductionOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []ProductionOrder
	for _, o := range s.orders {
		if statusFilter != "" && o.Status != statusFilter {
			continue
		}
		out = append(out, *o)
	}
	return out
}

// Update updates an existing order. Returns false if not found.
func (s *Store) Update(order *ProductionOrder) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.orders[order.ID]; !ok {
		return false
	}
	cp := *order
	s.orders[order.ID] = &cp
	return true
}
