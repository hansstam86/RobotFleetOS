package wms

import (
	"sync"
	"sync/atomic"
	"time"
)

// Store is an in-memory store for locations, inventory, and tasks.
type Store struct {
	mu         sync.RWMutex
	locations  map[string]*Location
	inventory  []Inventory // multiple rows per location+sku+lot
	tasks      map[string]*Task
	taskSeq    atomic.Uint64
}

// NewStore returns a new in-memory WMS store.
func NewStore() *Store {
	s := &Store{
		locations: make(map[string]*Location),
		inventory: make([]Inventory, 0),
		tasks:     make(map[string]*Task),
	}
	// Seed default receiving and staging for demo
	s.locations["RECV-01"] = &Location{ID: "RECV-01", ZoneID: "area-1", Type: LocationTypeReceiving, Name: "Receiving dock 1"}
	s.locations["STAGE-01"] = &Location{ID: "STAGE-01", ZoneID: "area-1", Type: LocationTypeStaging, Name: "Staging 1"}
	s.locations["A-01-01"] = &Location{ID: "A-01-01", ZoneID: "area-1", Type: LocationTypeStorage, Name: "Aisle 1, bin 1"}
	s.locations["A-01-02"] = &Location{ID: "A-01-02", ZoneID: "area-1", Type: LocationTypeStorage, Name: "Aisle 1, bin 2"}
	return s
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

// CreateLocation adds a location.
func (s *Store) CreateLocation(loc *Location) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if loc.ID == "" {
		return
	}
	cp := *loc
	s.locations[loc.ID] = &cp
}

// ListLocations returns all locations.
func (s *Store) ListLocations() []Location {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Location, 0, len(s.locations))
	for _, loc := range s.locations {
		out = append(out, *loc)
	}
	return out
}

// GetLocation returns a location by ID.
func (s *Store) GetLocation(id string) *Location {
	s.mu.RLock()
	defer s.mu.RUnlock()
	loc, ok := s.locations[id]
	if !ok {
		return nil
	}
	cp := *loc
	return &cp
}

// SetInventory sets or adjusts inventory at a location (by location+sku+lot).
func (s *Store) SetInventory(locationID, sku string, quantity int, lot string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	found := false
	for i := range s.inventory {
		if s.inventory[i].LocationID == locationID && s.inventory[i].SKU == sku && s.inventory[i].Lot == lot {
			s.inventory[i].Quantity = quantity
			s.inventory[i].UpdatedAt = now
			found = true
			break
		}
	}
	if !found {
		s.inventory = append(s.inventory, Inventory{
			LocationID: locationID,
			SKU:        sku,
			Quantity:   quantity,
			Lot:        lot,
			UpdatedAt:  now,
		})
	}
}

// AddInventory adds quantity to existing inventory (or creates). Use negative to subtract.
func (s *Store) AddInventory(locationID, sku string, delta int, lot string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	for i := range s.inventory {
		if s.inventory[i].LocationID == locationID && s.inventory[i].SKU == sku && s.inventory[i].Lot == lot {
			s.inventory[i].Quantity += delta
			if s.inventory[i].Quantity < 0 {
				s.inventory[i].Quantity = 0
			}
			s.inventory[i].UpdatedAt = now
			return
		}
	}
	if delta > 0 {
		s.inventory = append(s.inventory, Inventory{
			LocationID: locationID,
			SKU:        sku,
			Quantity:   delta,
			Lot:        lot,
			UpdatedAt:  now,
		})
	}
}

// ListInventory returns inventory, optionally filtered by location and/or sku.
func (s *Store) ListInventory(locationID, sku string) []Inventory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Inventory
	for _, inv := range s.inventory {
		if inv.Quantity <= 0 {
			continue
		}
		if locationID != "" && inv.LocationID != locationID {
			continue
		}
		if sku != "" && inv.SKU != sku {
			continue
		}
		out = append(out, inv)
	}
	return out
}

// CreateTask adds a task and returns it with ID set.
func (s *Store) CreateTask(t *Task) *Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t.ID == "" {
		t.ID = "wms-" + formatSeq(s.taskSeq.Add(1))
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	if t.Status == "" {
		t.Status = TaskStatusPending
	}
	cp := *t
	s.tasks[t.ID] = &cp
	return &cp
}

// GetTask returns a task by ID.
func (s *Store) GetTask(id string) *Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil
	}
	cp := *t
	return &cp
}

// UpdateTask updates a task.
func (s *Store) UpdateTask(t *Task) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[t.ID]; !ok {
		return false
	}
	cp := *t
	s.tasks[t.ID] = &cp
	return true
}

// ListTasks returns tasks, optionally filtered by status and type.
func (s *Store) ListTasks(status TaskStatus, taskType TaskType) []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Task
	for _, t := range s.tasks {
		if status != "" && t.Status != status {
			continue
		}
		if taskType != "" && t.Type != taskType {
			continue
		}
		out = append(out, *t)
	}
	return out
}
