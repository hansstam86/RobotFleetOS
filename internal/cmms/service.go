package cmms

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Service implements CMMS business logic.
type Service struct {
	Store        *Store
	FleetClient  *FleetClient
}

// NewService returns a CMMS service.
func NewService(store *Store, fleet *FleetClient) *Service {
	return &Service{Store: store, FleetClient: fleet}
}

// CreateEquipment creates equipment.
func (s *Service) CreateEquipment(ctx context.Context, e *Equipment) (*Equipment, error) {
	if e.Name == "" {
		return nil, fmt.Errorf("name required")
	}
	if e.Type == "" {
		e.Type = EquipmentTypeOther
	}
	return s.Store.CreateEquipment(e), nil
}

// GetEquipment returns equipment by ID.
func (s *Service) GetEquipment(ctx context.Context, id string) *Equipment {
	return s.Store.GetEquipment(id)
}

// UpdateEquipment updates equipment.
func (s *Service) UpdateEquipment(ctx context.Context, e *Equipment) (*Equipment, error) {
	if s.Store.GetEquipment(e.ID) == nil {
		return nil, fmt.Errorf("equipment not found: %s", e.ID)
	}
	s.Store.UpdateEquipment(e)
	return s.Store.GetEquipment(e.ID), nil
}

// ListEquipment returns equipment with optional filters.
func (s *Service) ListEquipment(ctx context.Context, status EquipmentStatus, eqType EquipmentType) []Equipment {
	return s.Store.ListEquipment(status, eqType)
}

// CreateMWO creates a maintenance work order.
func (s *Service) CreateMWO(ctx context.Context, m *MaintenanceWorkOrder) (*MaintenanceWorkOrder, error) {
	if m.EquipmentID == "" {
		return nil, fmt.Errorf("equipment_id required")
	}
	if s.Store.GetEquipment(m.EquipmentID) == nil {
		return nil, fmt.Errorf("equipment not found: %s", m.EquipmentID)
	}
	if m.Type == "" {
		m.Type = MWOCorrective
	}
	if m.Priority == 0 {
		m.Priority = 3
	}
	return s.Store.CreateMWO(m), nil
}

// GetMWO returns an MWO by ID.
func (s *Service) GetMWO(ctx context.Context, id string) *MaintenanceWorkOrder {
	return s.Store.GetMWO(id)
}

// StartMWO sets MWO status to in_progress and optionally sets equipment to under_maintenance.
func (s *Service) StartMWO(ctx context.Context, id string) (*MaintenanceWorkOrder, error) {
	m := s.Store.GetMWO(id)
	if m == nil {
		return nil, fmt.Errorf("mwo not found: %s", id)
	}
	if m.Status != MWOStatusOpen {
		return nil, fmt.Errorf("mwo not open: %s", m.Status)
	}
	m.Status = MWOStatusInProgress
	s.Store.UpdateMWO(m)
	if e := s.Store.GetEquipment(m.EquipmentID); e != nil {
		e.Status = EquipmentUnderMaintenance
		s.Store.UpdateEquipment(e)
	}
	return s.Store.GetMWO(id), nil
}

// CompleteMWO sets MWO status to completed and equipment back to operational.
func (s *Service) CompleteMWO(ctx context.Context, id string) (*MaintenanceWorkOrder, error) {
	m := s.Store.GetMWO(id)
	if m == nil {
		return nil, fmt.Errorf("mwo not found: %s", id)
	}
	if m.Status != MWOStatusInProgress && m.Status != MWOStatusOpen {
		return nil, fmt.Errorf("mwo cannot be completed: %s", m.Status)
	}
	now := time.Now().UTC()
	m.Status = MWOStatusCompleted
	m.CompletedAt = &now
	s.Store.UpdateMWO(m)
	if e := s.Store.GetEquipment(m.EquipmentID); e != nil {
		e.Status = EquipmentOperational
		s.Store.UpdateEquipment(e)
	}
	return s.Store.GetMWO(id), nil
}

// CancelMWO sets MWO status to cancelled.
func (s *Service) CancelMWO(ctx context.Context, id string) (*MaintenanceWorkOrder, error) {
	m := s.Store.GetMWO(id)
	if m == nil {
		return nil, fmt.Errorf("mwo not found: %s", id)
	}
	if m.Status == MWOStatusCompleted || m.Status == MWOStatusCancelled {
		return nil, fmt.Errorf("mwo already %s", m.Status)
	}
	m.Status = MWOStatusCancelled
	s.Store.UpdateMWO(m)
	if e := s.Store.GetEquipment(m.EquipmentID); e != nil && e.Status == EquipmentUnderMaintenance {
		e.Status = EquipmentOperational
		s.Store.UpdateEquipment(e)
	}
	return s.Store.GetMWO(id), nil
}

// ListMWOs returns MWOs with optional filters.
func (s *Service) ListMWOs(ctx context.Context, status MWOStatus, equipmentID string) []MaintenanceWorkOrder {
	return s.Store.ListMWOs(status, equipmentID)
}

// SubmitMWOToFleet sends the MWO to Fleet as a work order. Uses equipment's area_id; if missing, uses "maintenance".
func (s *Service) SubmitMWOToFleet(ctx context.Context, mwoID string, priority int) (fleetWorkOrderID string, err error) {
	m := s.Store.GetMWO(mwoID)
	if m == nil {
		return "", fmt.Errorf("mwo not found: %s", mwoID)
	}
	if m.FleetWorkOrderID != "" {
		return "", fmt.Errorf("mwo already submitted to fleet: %s", m.FleetWorkOrderID)
	}
	e := s.Store.GetEquipment(m.EquipmentID)
	areaID := "maintenance"
	if e != nil && e.AreaID != "" {
		areaID = e.AreaID
	}
	payload := map[string]interface{}{
		"type":               "maintenance",
		"cmms_work_order_id": mwoID,
		"equipment_id":       m.EquipmentID,
		"description":        m.Description,
		"maintenance_type":   string(m.Type),
	}
	if e != nil {
		payload["equipment_name"] = e.Name
		payload["zone_id"] = e.ZoneID
	}
	if m.Type == MWOFirmwareUpgrade && m.TargetFirmwareVersion != "" {
		payload["target_firmware_version"] = m.TargetFirmwareVersion
	}
	payloadBytes, _ := json.Marshal(payload)
	var payloadMap map[string]interface{}
	_ = json.Unmarshal(payloadBytes, &payloadMap)

	fid, err := s.FleetClient.SubmitMaintenanceWorkOrder(ctx, areaID, priority, payloadMap)
	if err != nil {
		return "", err
	}
	m.FleetWorkOrderID = fid
	s.Store.UpdateMWO(m)
	return fid, nil
}

// TriggerFirmwareCampaign triggers a firmware campaign on Fleet (broadcast to zone; busy robots defer).
func (s *Service) TriggerFirmwareCampaign(ctx context.Context, seedBusy int) (message string, err error) {
	return s.FleetClient.TriggerFirmwareSimulate(ctx, seedBusy)
}
