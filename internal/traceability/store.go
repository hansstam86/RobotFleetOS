package traceability

import (
	"sync"
	"sync/atomic"
	"time"
)

// Store is an in-memory store for trace records, indexed by serial and lot.
type Store struct {
	mu       sync.RWMutex
	records  []TraceRecord
	bySerial map[string][]int // serial -> indices into records
	byLot    map[string][]int // lot -> indices into records
	seq      atomic.Uint64
}

// NewStore returns a new traceability store.
func NewStore() *Store {
	return &Store{
		records:  make([]TraceRecord, 0),
		bySerial: make(map[string][]int),
		byLot:    make(map[string][]int),
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

// Add appends a record and indexes it by serial and lot.
func (s *Store) Add(r *TraceRecord) *TraceRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.ID == "" {
		r.ID = "tr-" + formatSeq(s.seq.Add(1))
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	cp := *r
	if cp.Extra != nil {
		extra := make(map[string]string, len(cp.Extra))
		for k, v := range cp.Extra {
			extra[k] = v
		}
		cp.Extra = extra
	}
	idx := len(s.records)
	s.records = append(s.records, cp)
	if cp.Serial != "" {
		s.bySerial[cp.Serial] = append(s.bySerial[cp.Serial], idx)
	}
	if cp.Lot != "" {
		s.byLot[cp.Lot] = append(s.byLot[cp.Lot], idx)
	}
	return &s.records[idx]
}

// BySerial returns all records for the given serial (chronological order).
func (s *Store) BySerial(serial string) []TraceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	idxs, ok := s.bySerial[serial]
	if !ok {
		return nil
	}
	out := make([]TraceRecord, 0, len(idxs))
	for _, i := range idxs {
		if i < len(s.records) {
			cp := s.records[i]
			out = append(out, cp)
		}
	}
	return out
}

// ByLot returns all records for the given lot (chronological order).
func (s *Store) ByLot(lot string) []TraceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	idxs, ok := s.byLot[lot]
	if !ok {
		return nil
	}
	out := make([]TraceRecord, 0, len(idxs))
	for _, i := range idxs {
		if i < len(s.records) {
			cp := s.records[i]
			out = append(out, cp)
		}
	}
	return out
}

// Recall returns records matching filters (lot, sku, from_date, to_date) for recall reporting.
func (s *Store) Recall(lot, sku string, from, to *time.Time) []TraceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []TraceRecord
	for i := range s.records {
		r := s.records[i]
		if lot != "" && r.Lot != lot {
			continue
		}
		if sku != "" && r.SKU != sku {
			continue
		}
		if from != nil && r.CreatedAt.Before(*from) {
			continue
		}
		if to != nil && r.CreatedAt.After(*to) {
			continue
		}
		out = append(out, r)
	}
	return out
}

// Count returns total record count (for dashboard).
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.records)
}
