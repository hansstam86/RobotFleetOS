package erp

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Store struct {
	mu     sync.RWMutex
	seq    atomic.Uint64
	orders map[string]*Order
}

func NewStore() *Store {
	return &Store{orders: make(map[string]*Order)}
}

func seqID(prefix string, n uint64) string { return prefix + "-" + fmt.Sprintf("%d", n) }

func (s *Store) CreateOrder(o *Order) *Order {
	s.mu.Lock()
	defer s.mu.Unlock()
	if o.ID == "" {
		o.ID = seqID("erp", s.seq.Add(1))
	}
	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now().UTC()
	}
	if o.Status == "" {
		o.Status = OrderStatusDraft
	}
	cp := *o
	s.orders[cp.ID] = &cp
	return &cp
}

func (s *Store) GetOrder(id string) *Order {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	if !ok {
		return nil
	}
	cp := *o
	return &cp
}

func (s *Store) UpdateOrder(o *Order) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.orders[o.ID]; !ok {
		return false
	}
	cp := *o
	s.orders[o.ID] = &cp
	return true
}

func (s *Store) ListOrders(status OrderStatus) []Order {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Order
	for _, o := range s.orders {
		if status != "" && o.Status != status {
			continue
		}
		out = append(out, *o)
	}
	return out
}
