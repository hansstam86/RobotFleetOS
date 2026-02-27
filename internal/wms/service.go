package wms

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Service implements WMS business logic.
type Service struct {
	Store            *Store
	FleetClient      *FleetClient
	WarehouseAreaID  string // area_id sent to Fleet for warehouse tasks (e.g. area-1)
}

// NewService returns a WMS service.
func NewService(store *Store, fleet *FleetClient, warehouseAreaID string) *Service {
	if warehouseAreaID == "" {
		warehouseAreaID = "area-1"
	}
	return &Service{Store: store, FleetClient: fleet, WarehouseAreaID: warehouseAreaID}
}

// CreateLocation adds a location.
func (s *Service) CreateLocation(ctx context.Context, loc *Location) error {
	if loc.ID == "" {
		return fmt.Errorf("location id required")
	}
	s.Store.CreateLocation(loc)
	return nil
}

// ListLocations returns all locations.
func (s *Service) ListLocations(ctx context.Context) []Location {
	return s.Store.ListLocations()
}

// GetLocation returns a location by ID.
func (s *Service) GetLocation(ctx context.Context, id string) *Location {
	return s.Store.GetLocation(id)
}

// ReceiveInventory adds quantity at a location (e.g. receiving).
func (s *Service) ReceiveInventory(ctx context.Context, locationID, sku string, quantity int, lot string) error {
	if locationID == "" || sku == "" {
		return fmt.Errorf("location_id and sku required")
	}
	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	s.Store.AddInventory(locationID, sku, quantity, lot)
	return nil
}

// ListInventory returns inventory, optionally filtered.
func (s *Service) ListInventory(ctx context.Context, locationID, sku string) []Inventory {
	return s.Store.ListInventory(locationID, sku)
}

// CreateTask creates a pick, putaway, or move task.
func (s *Service) CreateTask(ctx context.Context, t *Task) (*Task, error) {
	if t.Type == "" || t.SKU == "" || t.Quantity <= 0 {
		return nil, fmt.Errorf("type, sku, and quantity required")
	}
	switch t.Type {
	case TaskTypePick:
		if t.FromLocationID == "" {
			return nil, fmt.Errorf("from_location_id required for pick")
		}
	case TaskTypePutaway:
		if t.FromLocationID == "" || t.ToLocationID == "" {
			return nil, fmt.Errorf("from_location_id and to_location_id required for putaway")
		}
	case TaskTypeMove:
		if t.FromLocationID == "" || t.ToLocationID == "" {
			return nil, fmt.Errorf("from_location_id and to_location_id required for move")
		}
	default:
		return nil, fmt.Errorf("invalid task type: %s", t.Type)
	}
	return s.Store.CreateTask(t), nil
}

// GetTask returns a task by ID.
func (s *Service) GetTask(ctx context.Context, id string) *Task {
	return s.Store.GetTask(id)
}

// ListTasks returns tasks with optional filters.
func (s *Service) ListTasks(ctx context.Context, status TaskStatus, taskType TaskType) []Task {
	return s.Store.ListTasks(status, taskType)
}

// ReleaseTask submits the task to Fleet and marks it released.
func (s *Service) ReleaseTask(ctx context.Context, id string) (*Task, error) {
	t := s.Store.GetTask(id)
	if t == nil {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	if t.Status != TaskStatusPending {
		return nil, fmt.Errorf("task cannot be released from status %s", t.Status)
	}
	payload := map[string]interface{}{
		"wms_task_id":       t.ID,
		"type":              string(t.Type),
		"from_location_id":  t.FromLocationID,
		"to_location_id":    t.ToLocationID,
		"sku":               t.SKU,
		"quantity":          t.Quantity,
		"lot":               t.Lot,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	fleetID, err := s.FleetClient.SubmitWorkOrder(ctx, s.WarehouseAreaID, 1, payloadBytes)
	if err != nil {
		return nil, fmt.Errorf("submit to fleet: %w", err)
	}
	now := time.Now().UTC()
	t.FleetWorkOrderID = fleetID
	t.Status = TaskStatusInProgress
	t.ReleasedAt = &now
	s.Store.UpdateTask(t)
	return s.Store.GetTask(id), nil
}

// CompleteTask marks the task completed and updates inventory.
func (s *Service) CompleteTask(ctx context.Context, id string) (*Task, error) {
	t := s.Store.GetTask(id)
	if t == nil {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	if t.Status == TaskStatusCompleted || t.Status == TaskStatusCancelled {
		return nil, fmt.Errorf("task already terminal")
	}
	now := time.Now().UTC()
	t.Status = TaskStatusCompleted
	t.CompletedAt = &now
	switch t.Type {
	case TaskTypePick:
		s.Store.AddInventory(t.FromLocationID, t.SKU, -t.Quantity, t.Lot)
	case TaskTypePutaway:
		s.Store.AddInventory(t.FromLocationID, t.SKU, -t.Quantity, t.Lot)
		s.Store.AddInventory(t.ToLocationID, t.SKU, t.Quantity, t.Lot)
	case TaskTypeMove:
		s.Store.AddInventory(t.FromLocationID, t.SKU, -t.Quantity, t.Lot)
		s.Store.AddInventory(t.ToLocationID, t.SKU, t.Quantity, t.Lot)
	}
	s.Store.UpdateTask(t)
	return s.Store.GetTask(id), nil
}

// CancelTask marks the task cancelled.
func (s *Service) CancelTask(ctx context.Context, id string) (*Task, error) {
	t := s.Store.GetTask(id)
	if t == nil {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	if t.Status == TaskStatusCompleted || t.Status == TaskStatusCancelled {
		return nil, fmt.Errorf("task already terminal")
	}
	t.Status = TaskStatusCancelled
	s.Store.UpdateTask(t)
	return s.Store.GetTask(id), nil
}
