package wms

import "time"

// TaskType is the type of warehouse task.
type TaskType string

const (
	TaskTypePick    TaskType = "pick"
	TaskTypePutaway TaskType = "putaway"
	TaskTypeMove    TaskType = "move"
)

// TaskStatus is the lifecycle state of a warehouse task.
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusReleased   TaskStatus = "released"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// LocationType describes the purpose of a location.
type LocationType string

const (
	LocationTypeReceiving LocationType = "receiving"
	LocationTypeStorage   LocationType = "storage"
	LocationTypeStaging   LocationType = "staging"
	LocationTypeShipping  LocationType = "shipping"
)

// Location is a storage location in the warehouse.
type Location struct {
	ID     string       `json:"id"`
	ZoneID string       `json:"zone_id"`
	Type   LocationType `json:"type"`
	Name   string       `json:"name,omitempty"`
}

// Inventory is quantity of a SKU at a location.
type Inventory struct {
	LocationID string    `json:"location_id"`
	SKU       string    `json:"sku"`
	Quantity  int       `json:"quantity"`
	Lot       string    `json:"lot,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Task is a warehouse task (pick, putaway, move).
type Task struct {
	ID               string     `json:"id"`
	Type             TaskType   `json:"type"`
	Status           TaskStatus `json:"status"`
	FromLocationID   string     `json:"from_location_id,omitempty"`
	ToLocationID     string     `json:"to_location_id,omitempty"`
	SKU              string     `json:"sku"`
	Quantity         int        `json:"quantity"`
	Lot              string     `json:"lot,omitempty"`
	FleetWorkOrderID string     `json:"fleet_work_order_id,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	ReleasedAt       *time.Time `json:"released_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}
