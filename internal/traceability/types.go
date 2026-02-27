package traceability

import "time"

// EventType is the type of traceability event.
type EventType string

const (
	EventProduced        EventType = "produced"
	EventReceived        EventType = "received"
	EventShipped         EventType = "shipped"
	EventComponentLinked EventType = "component_linked"
	EventInspection      EventType = "inspection"
	EventRework           EventType = "rework"
	EventScrap            EventType = "scrap"
)

// TraceRecord is a single traceability event (serial/lot at a step).
type TraceRecord struct {
	ID          string            `json:"id"`
	Serial      string            `json:"serial,omitempty"`
	Lot         string            `json:"lot,omitempty"`
	EventType   EventType         `json:"event_type"`
	SKU         string            `json:"sku"`
	Quantity    int               `json:"quantity,omitempty"`
	MESOrderID  string            `json:"mes_order_id,omitempty"`
	WMSTaskID   string            `json:"wms_task_id,omitempty"`
	FleetWorkOrderID string       `json:"fleet_work_order_id,omitempty"`
	StationID   string            `json:"station_id,omitempty"`
	ZoneID      string            `json:"zone_id,omitempty"`
	ParentSerial string           `json:"parent_serial,omitempty"` // assembly: this serial is part of parent
	CreatedAt   time.Time         `json:"created_at"`
	Extra       map[string]string `json:"extra,omitempty"`
}
