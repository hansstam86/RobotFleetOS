package qms

import (
	"sync"
	"sync/atomic"
	"time"
)

// Store is an in-memory store for inspections, NCRs, and holds.
type Store struct {
	mu          sync.RWMutex
	inspections []Inspection
	insSeq      atomic.Uint64
	ncrSeq      atomic.Uint64
	ncrMap      map[string]*NCR
	holdSeq     atomic.Uint64
	holdMap     map[string]*Hold
}

// NewStore returns a new QMS store.
func NewStore() *Store {
	return &Store{
		inspections: make([]Inspection, 0),
		ncrMap:      make(map[string]*NCR),
		holdMap:     make(map[string]*Hold),
	}
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

// AddInspection appends an inspection.
func (s *Store) AddInspection(r *Inspection) *Inspection {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.ID == "" {
		r.ID = "ins-" + formatSeq(s.insSeq.Add(1))
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	cp := *r
	s.inspections = append(s.inspections, cp)
	return &s.inspections[len(s.inspections)-1]
}

// ListInspections returns inspections, optionally filtered by serial or lot.
func (s *Store) ListInspections(serial, lot string) []Inspection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Inspection
	for _, r := range s.inspections {
		if serial != "" && r.Serial != serial {
			continue
		}
		if lot != "" && r.Lot != lot {
			continue
		}
		out = append(out, r)
	}
	return out
}

// CreateNCR adds an NCR.
func (s *Store) CreateNCR(n *NCR) *NCR {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n.ID == "" {
		n.ID = "ncr-" + formatSeq(s.ncrSeq.Add(1))
	}
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now().UTC()
	}
	if n.Status == "" {
		n.Status = NCRStatusOpen
	}
	cp := *n
	s.ncrMap[cp.ID] = &cp
	return &cp
}

// GetNCR returns an NCR by ID.
func (s *Store) GetNCR(id string) *NCR {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.ncrMap[id]
	if !ok {
		return nil
	}
	cp := *n
	return &cp
}

// UpdateNCR updates an NCR.
func (s *Store) UpdateNCR(n *NCR) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.ncrMap[n.ID]; !ok {
		return false
	}
	cp := *n
	s.ncrMap[n.ID] = &cp
	return true
}

// ListNCRs returns NCRs, optionally filtered by status.
func (s *Store) ListNCRs(status NCRStatus) []NCR {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []NCR
	for _, n := range s.ncrMap {
		if status != "" && n.Status != status {
			continue
		}
		out = append(out, *n)
	}
	return out
}

// CreateHold adds a hold.
func (s *Store) CreateHold(h *Hold) *Hold {
	s.mu.Lock()
	defer s.mu.Unlock()
	if h.ID == "" {
		h.ID = "hold-" + formatSeq(s.holdSeq.Add(1))
	}
	if h.HeldAt.IsZero() {
		h.HeldAt = time.Now().UTC()
	}
	cp := *h
	s.holdMap[cp.ID] = &cp
	return &cp
}

// GetHold returns a hold by ID.
func (s *Store) GetHold(id string) *Hold {
	s.mu.RLock()
	defer s.mu.RUnlock()
	h, ok := s.holdMap[id]
	if !ok {
		return nil
	}
	cp := *h
	return &cp
}

// UpdateHold updates a hold.
func (s *Store) UpdateHold(h *Hold) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.holdMap[h.ID]; !ok {
		return false
	}
	cp := *h
	s.holdMap[h.ID] = &cp
	return true
}

// ListHolds returns holds (active only if activeOnly).
func (s *Store) ListHolds(activeOnly bool) []Hold {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Hold
	for _, h := range s.holdMap {
		if activeOnly && h.ReleasedAt != nil {
			continue
		}
		out = append(out, *h)
	}
	return out
}
