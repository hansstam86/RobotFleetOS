package mes

import "time"

// OrderStatus is the lifecycle state of a production order.
type OrderStatus string

const (
	OrderStatusDraft      OrderStatus = "draft"
	OrderStatusReleased   OrderStatus = "released"
	OrderStatusInProgress OrderStatus = "in_progress"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusPaused     OrderStatus = "paused"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

// ProductionOrder is a manufacturing order tracked by MES (created from ERP or internally).
type ProductionOrder struct {
	ID                 string       `json:"id"`
	ERPOrderRef        string       `json:"erp_order_ref,omitempty"`
	SKU                string       `json:"sku"`
	ProductID          string       `json:"product_id,omitempty"` // alias for SKU or variant
	Quantity           int          `json:"quantity"`
	QuantityCompleted  int          `json:"quantity_completed"`
	QuantityScrapped   int          `json:"quantity_scrapped"`
	AreaID             string       `json:"area_id"`
	ZoneID             string       `json:"zone_id,omitempty"`
	Status             OrderStatus  `json:"status"`
	Priority           int          `json:"priority"`
	BOMRevision        string       `json:"bom_revision,omitempty"`
	RoutingRevision    string       `json:"routing_revision,omitempty"`
	FleetWorkOrderID   string       `json:"fleet_work_order_id,omitempty"` // set when released to Fleet
	CreatedAt          time.Time    `json:"created_at"`
	ReleasedAt         *time.Time   `json:"released_at,omitempty"`
	CompletedAt        *time.Time   `json:"completed_at,omitempty"`
	ScrapRecords       []ScrapRecord `json:"scrap_records,omitempty"`
}

// ScrapRecord is a single scrap report for an order.
type ScrapRecord struct {
	Quantity   int       `json:"quantity"`
	Reason     string    `json:"reason"`
	RecordedAt time.Time `json:"recorded_at"`
}
