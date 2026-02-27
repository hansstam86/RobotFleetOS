package traceability

import (
	"context"
	"fmt"
	"time"
)

// Service implements traceability business logic.
type Service struct {
	Store *Store
}

// NewService returns a traceability service.
func NewService(store *Store) *Service {
	return &Service{Store: store}
}

// Record adds a trace record. At least one of serial or lot should be set.
func (s *Service) Record(ctx context.Context, r *TraceRecord) (*TraceRecord, error) {
	if r.EventType == "" {
		return nil, fmt.Errorf("event_type required")
	}
	if r.SKU == "" {
		return nil, fmt.Errorf("sku required")
	}
	if r.Serial == "" && r.Lot == "" {
		return nil, fmt.Errorf("serial or lot required")
	}
	if r.Quantity <= 0 && r.EventType != EventScrap {
		r.Quantity = 1
	}
	return s.Store.Add(r), nil
}

// GetGenealogyBySerial returns all records for the given serial.
func (s *Service) GetGenealogyBySerial(ctx context.Context, serial string) []TraceRecord {
	return s.Store.BySerial(serial)
}

// GetGenealogyByLot returns all records for the given lot.
func (s *Service) GetGenealogyByLot(ctx context.Context, lot string) []TraceRecord {
	return s.Store.ByLot(lot)
}

// Recall returns records matching filters for recall reporting.
func (s *Service) Recall(ctx context.Context, lot, sku string, from, to *time.Time) []TraceRecord {
	return s.Store.Recall(lot, sku, from, to)
}

// RecordCount returns total number of trace records.
func (s *Service) RecordCount(ctx context.Context) int {
	return s.Store.Count()
}
