package cmms

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Store is an in-memory store for equipment and maintenance work orders.
type Store struct {
	mu       sync.RWMutex
	equipSeq atomic.Uint64
	equip    map[string]*Equipment
	mwoSeq   atomic.Uint64
	mwo      map[string]*MaintenanceWorkOrder
}

// NewStore returns a new CMMS store.
func NewStore() *Store {
	return &Store{
		equip: make(map[string]*Equipment),
		mwo:   make(map[string]*MaintenanceWorkOrder),
	}
}

func seqID(prefix string, n uint64) string {
	return prefix + "-" + fmt.Sprintf("%d", n)
}

// CreateEquipment adds equipment. ID is set if empty.
func (s *Store) CreateEquipment(e *Equipment) *Equipment {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e.ID == "" {
		e.ID = seqID("eq", s.equipSeq.Add(1))
	}
	now := time.Now().UTC()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	e.UpdatedAt = now
	if e.Status == "" {
		e.Status = EquipmentOperational
	}
	cp := *e
	s.equip[cp.ID] = &cp
	return &cp
}

// GetEquipment returns equipment by ID.
func (s *Store) GetEquipment(id string) *Equipment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.equip[id]
	if !ok {
		return nil
	}
	cp := *e
	return &cp
}

// UpdateEquipment updates equipment.
func (s *Store) UpdateEquipment(e *Equipment) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.equip[e.ID]; !ok {
		return false
	}
	e.UpdatedAt = time.Now().UTC()
	cp := *e
	s.equip[e.ID] = &cp
	return true
}

// ListEquipment returns all equipment, optionally filtered by status or type.
func (s *Store) ListEquipment(status EquipmentStatus, eqType EquipmentType) []Equipment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Equipment
	for _, e := range s.equip {
		if status != "" && e.Status != status {
			continue
		}
		if eqType != "" && e.Type != eqType {
			continue
		}
		out = append(out, *e)
	}
	return out
}

// CreateMWO adds a maintenance work order.
func (s *Store) CreateMWO(m *MaintenanceWorkOrder) *MaintenanceWorkOrder {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m.ID == "" {
		m.ID = seqID("mwo", s.mwoSeq.Add(1))
	}
	now := time.Now().UTC()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
	if m.Status == "" {
		m.Status = MWOStatusOpen
	}
	cp := *m
	s.mwo[cp.ID] = &cp
	return &cp
}

// GetMWO returns an MWO by ID.
func (s *Store) GetMWO(id string) *MaintenanceWorkOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.mwo[id]
	if !ok {
		return nil
	}
	cp := *m
	return &cp
}

// UpdateMWO updates an MWO.
func (s *Store) UpdateMWO(m *MaintenanceWorkOrder) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.mwo[m.ID]; !ok {
		return false
	}
	m.UpdatedAt = time.Now().UTC()
	cp := *m
	s.mwo[m.ID] = &cp
	return true
}

// ListMWOs returns MWOs, optionally filtered by status or equipment_id.
func (s *Store) ListMWOs(status MWOStatus, equipmentID string) []MaintenanceWorkOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []MaintenanceWorkOrder
	for _, m := range s.mwo {
		if status != "" && m.Status != status {
			continue
		}
		if equipmentID != "" && m.EquipmentID != equipmentID {
			continue
		}
		out = append(out, *m)
	}
	return out
}
