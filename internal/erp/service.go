package erp

import (
	"context"
	"fmt"
	"time"
)

type Service struct {
	Store         *Store
	MESClient     *MESClient
	DefaultAreaID string
}

func NewService(store *Store, mes *MESClient, defaultAreaID string) *Service {
	if defaultAreaID == "" {
		defaultAreaID = "area-1"
	}
	return &Service{Store: store, MESClient: mes, DefaultAreaID: defaultAreaID}
}

func (s *Service) CreateOrder(ctx context.Context, o *Order) (*Order, error) {
	if o.OrderRef == "" {
		return nil, fmt.Errorf("order_ref required")
	}
	if o.SKU == "" {
		return nil, fmt.Errorf("sku required")
	}
	if o.Quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	return s.Store.CreateOrder(o), nil
}

func (s *Service) GetOrder(ctx context.Context, id string) *Order {
	return s.Store.GetOrder(id)
}

func (s *Service) ListOrders(ctx context.Context, status OrderStatus) []Order {
	return s.Store.ListOrders(status)
}

func (s *Service) SubmitToMES(ctx context.Context, orderID, zoneID string, priority int) (*Order, error) {
	o := s.Store.GetOrder(orderID)
	if o == nil {
		return nil, fmt.Errorf("order not found: %s", orderID)
	}
	if o.Status == OrderStatusSubmitted {
		return nil, fmt.Errorf("order already submitted to MES: %s", o.MESOrderID)
	}
	if o.Status == OrderStatusCancelled {
		return nil, fmt.Errorf("order is cancelled")
	}
	mesID, err := s.MESClient.CreateProductionOrder(ctx, o.OrderRef, o.SKU, o.Quantity, s.DefaultAreaID, zoneID, priority)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	o.Status = OrderStatusSubmitted
	o.MESOrderID = mesID
	o.SubmittedAt = &now
	s.Store.UpdateOrder(o)
	return s.Store.GetOrder(orderID), nil
}

func (s *Service) CancelOrder(ctx context.Context, id string) (*Order, error) {
	o := s.Store.GetOrder(id)
	if o == nil {
		return nil, fmt.Errorf("order not found: %s", id)
	}
	if o.Status != OrderStatusDraft {
		return nil, fmt.Errorf("only draft orders can be cancelled")
	}
	o.Status = OrderStatusCancelled
	s.Store.UpdateOrder(o)
	return s.Store.GetOrder(id), nil
}
