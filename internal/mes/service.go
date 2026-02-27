package mes

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Service implements MES business logic: create, release (to Fleet), pause, complete, scrap.
type Service struct {
	Store       *Store
	FleetClient *FleetClient
}

// NewService returns an MES service with the given store and Fleet client.
func NewService(store *Store, fleetClient *FleetClient) *Service {
	return &Service{Store: store, FleetClient: fleetClient}
}

// CreateOrder creates a new production order in draft status.
func (s *Service) CreateOrder(ctx context.Context, order *ProductionOrder) (*ProductionOrder, error) {
	if order.SKU == "" {
		return nil, fmt.Errorf("sku required")
	}
	if order.AreaID == "" {
		return nil, fmt.Errorf("area_id required")
	}
	if order.Quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	if order.Priority <= 0 {
		order.Priority = 1
	}
	return s.Store.Create(order)
}

// GetOrder returns an order by ID, or nil if not found.
func (s *Service) GetOrder(ctx context.Context, id string) *ProductionOrder {
	return s.Store.Get(id)
}

// ListOrders returns orders, optionally filtered by status.
func (s *Service) ListOrders(ctx context.Context, statusFilter OrderStatus) []ProductionOrder {
	return s.Store.List(statusFilter)
}

// ReleaseOrder submits the order to Fleet and marks it released/in_progress.
func (s *Service) ReleaseOrder(ctx context.Context, id string) (*ProductionOrder, error) {
	order := s.Store.Get(id)
	if order == nil {
		return nil, fmt.Errorf("order not found: %s", id)
	}
	if order.Status != OrderStatusDraft && order.Status != OrderStatusPaused {
		return nil, fmt.Errorf("order cannot be released from status %s", order.Status)
	}
	payload := map[string]interface{}{
		"mes_order_id":   order.ID,
		"erp_order_ref":  order.ERPOrderRef,
		"sku":            order.SKU,
		"product_id":     order.ProductID,
		"quantity":       order.Quantity,
		"bom_revision":   order.BOMRevision,
		"routing_revision": order.RoutingRevision,
	}
	if order.ZoneID != "" {
		payload["zone_id"] = order.ZoneID
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	fleetWOID, err := s.FleetClient.SubmitWorkOrder(ctx, order.AreaID, order.Priority, payloadBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("submit to fleet: %w", err)
	}
	now := time.Now().UTC()
	order.FleetWorkOrderID = fleetWOID
	order.Status = OrderStatusInProgress
	order.ReleasedAt = &now
	ok := s.Store.Update(order)
	if !ok {
		return nil, fmt.Errorf("store update failed")
	}
	return s.Store.Get(id), nil
}

// PauseOrder marks the order as paused. Fleet work order is not cancelled (future: call Fleet cancel).
func (s *Service) PauseOrder(ctx context.Context, id string) (*ProductionOrder, error) {
	order := s.Store.Get(id)
	if order == nil {
		return nil, fmt.Errorf("order not found: %s", id)
	}
	if order.Status != OrderStatusInProgress && order.Status != OrderStatusReleased {
		return nil, fmt.Errorf("order cannot be paused from status %s", order.Status)
	}
	order.Status = OrderStatusPaused
	ok := s.Store.Update(order)
	if !ok {
		return nil, fmt.Errorf("store update failed")
	}
	return s.Store.Get(id), nil
}

// CompleteOrder marks the order as completed.
func (s *Service) CompleteOrder(ctx context.Context, id string) (*ProductionOrder, error) {
	order := s.Store.Get(id)
	if order == nil {
		return nil, fmt.Errorf("order not found: %s", id)
	}
	if order.Status == OrderStatusCompleted || order.Status == OrderStatusCancelled {
		return nil, fmt.Errorf("order already terminal: %s", order.Status)
	}
	now := time.Now().UTC()
	order.Status = OrderStatusCompleted
	order.CompletedAt = &now
	if order.QuantityCompleted == 0 {
		order.QuantityCompleted = order.Quantity - order.QuantityScrapped
	}
	ok := s.Store.Update(order)
	if !ok {
		return nil, fmt.Errorf("store update failed")
	}
	return s.Store.Get(id), nil
}

// CancelOrder marks the order as cancelled.
func (s *Service) CancelOrder(ctx context.Context, id string) (*ProductionOrder, error) {
	order := s.Store.Get(id)
	if order == nil {
		return nil, fmt.Errorf("order not found: %s", id)
	}
	if order.Status == OrderStatusCompleted || order.Status == OrderStatusCancelled {
		return nil, fmt.Errorf("order already terminal: %s", order.Status)
	}
	order.Status = OrderStatusCancelled
	ok := s.Store.Update(order)
	if !ok {
		return nil, fmt.Errorf("store update failed")
	}
	return s.Store.Get(id), nil
}

// TriggerFirmwareUpdate calls Fleet's firmware simulate endpoint (for use during production).
func (s *Service) TriggerFirmwareUpdate(ctx context.Context, seedBusy int) (message string, err error) {
	return s.FleetClient.FirmwareSimulate(ctx, seedBusy)
}

// ReportScrap adds a scrap record and updates quantity scrapped.
func (s *Service) ReportScrap(ctx context.Context, id string, quantity int, reason string) (*ProductionOrder, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	order := s.Store.Get(id)
	if order == nil {
		return nil, fmt.Errorf("order not found: %s", id)
	}
	if order.Status == OrderStatusCompleted || order.Status == OrderStatusCancelled {
		return nil, fmt.Errorf("cannot report scrap for terminal order")
	}
	order.QuantityScrapped += quantity
	if order.ScrapRecords == nil {
		order.ScrapRecords = []ScrapRecord{}
	}
	order.ScrapRecords = append(order.ScrapRecords, ScrapRecord{
		Quantity:   quantity,
		Reason:     reason,
		RecordedAt: time.Now().UTC(),
	})
	ok := s.Store.Update(order)
	if !ok {
		return nil, fmt.Errorf("store update failed")
	}
	return s.Store.Get(id), nil
}
