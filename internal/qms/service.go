package qms

import (
	"context"
	"fmt"
	"time"
)

// Service implements QMS business logic.
type Service struct {
	Store *Store
}

// NewService returns a QMS service.
func NewService(store *Store) *Service {
	return &Service{Store: store}
}

// RecordInspection adds an inspection record.
func (s *Service) RecordInspection(ctx context.Context, r *Inspection) (*Inspection, error) {
	if r.Result != InspectionPass && r.Result != InspectionFail {
		return nil, fmt.Errorf("result must be pass or fail")
	}
	if r.SKU == "" {
		return nil, fmt.Errorf("sku required")
	}
	if r.Serial == "" && r.Lot == "" {
		return nil, fmt.Errorf("serial or lot required")
	}
	return s.Store.AddInspection(r), nil
}

// ListInspections returns inspections with optional filters.
func (s *Service) ListInspections(ctx context.Context, serial, lot string) []Inspection {
	return s.Store.ListInspections(serial, lot)
}

// CreateNCR creates a non-conformance report.
func (s *Service) CreateNCR(ctx context.Context, n *NCR) (*NCR, error) {
	if n.Description == "" {
		return nil, fmt.Errorf("description required")
	}
	if n.Serial == "" && n.Lot == "" {
		return nil, fmt.Errorf("serial or lot required")
	}
	return s.Store.CreateNCR(n), nil
}

// GetNCR returns an NCR by ID.
func (s *Service) GetNCR(ctx context.Context, id string) *NCR {
	return s.Store.GetNCR(id)
}

// CloseNCR marks an NCR as closed.
func (s *Service) CloseNCR(ctx context.Context, id string) (*NCR, error) {
	n := s.Store.GetNCR(id)
	if n == nil {
		return nil, fmt.Errorf("ncr not found: %s", id)
	}
	if n.Status == NCRStatusClosed {
		return nil, fmt.Errorf("ncr already closed")
	}
	now := time.Now().UTC()
	n.Status = NCRStatusClosed
	n.ClosedAt = &now
	s.Store.UpdateNCR(n)
	return s.Store.GetNCR(id), nil
}

// ListNCRs returns NCRs with optional status filter.
func (s *Service) ListNCRs(ctx context.Context, status NCRStatus) []NCR {
	return s.Store.ListNCRs(status)
}

// CreateHold places a hold on a serial or lot.
func (s *Service) CreateHold(ctx context.Context, h *Hold) (*Hold, error) {
	if h.Reason == "" {
		return nil, fmt.Errorf("reason required")
	}
	if h.Serial == "" && h.Lot == "" {
		return nil, fmt.Errorf("serial or lot required")
	}
	return s.Store.CreateHold(h), nil
}

// ReleaseHold releases a hold.
func (s *Service) ReleaseHold(ctx context.Context, id string) (*Hold, error) {
	h := s.Store.GetHold(id)
	if h == nil {
		return nil, fmt.Errorf("hold not found: %s", id)
	}
	if h.ReleasedAt != nil {
		return nil, fmt.Errorf("hold already released")
	}
	now := time.Now().UTC()
	h.ReleasedAt = &now
	s.Store.UpdateHold(h)
	return s.Store.GetHold(id), nil
}

// GetHold returns a hold by ID.
func (s *Service) GetHold(ctx context.Context, id string) *Hold {
	return s.Store.GetHold(id)
}

// ListHolds returns holds, optionally active only.
func (s *Service) ListHolds(ctx context.Context, activeOnly bool) []Hold {
	return s.Store.ListHolds(activeOnly)
}
