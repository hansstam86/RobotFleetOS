// Package api defines shared types and identifiers for RobotFleetOS.
package api

import "time"

// RobotID uniquely identifies a robot in the fleet (e.g. "area-1/zone-42/robot-007" or UUID).
type RobotID string

// ZoneID identifies a zone controller's domain.
type ZoneID string

// AreaID identifies an area controller's domain.
type AreaID string

// WorkOrderID identifies a fleet-level work order.
type WorkOrderID string

// TaskID identifies a zone-level or robot-level task.
type TaskID string

// WorkOrder is a high-level job from the fleet (e.g. "pick 10k items from A to B").
type WorkOrder struct {
	ID        WorkOrderID  `json:"id"`
	AreaID    AreaID       `json:"area_id"`
	Priority  int          `json:"priority"`
	Payload   []byte       `json:"payload"` // JSON or structured; schema defined by fleet
	CreatedAt time.Time    `json:"created_at"`
	Deadline  *time.Time   `json:"deadline,omitempty"`
}

// ZoneTask is a decomposed task assigned to a zone.
type ZoneTask struct {
	ID        TaskID      `json:"id"`
	ZoneID    ZoneID      `json:"zone_id"`
	OrderID   WorkOrderID `json:"order_id"`
	Payload   []byte      `json:"payload"`
	CreatedAt time.Time   `json:"created_at"`
}

// RobotCommand is a command sent from zone/edge to a robot (or edge gateway).
type RobotCommand struct {
	ID        TaskID    `json:"id"`
	RobotID   RobotID   `json:"robot_id"`
	Type      string    `json:"type"` // e.g. "MOVE", "PICK", "STOP"
	Payload   []byte    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

// RobotStatus is telemetry/heartbeat from edge to zone.
type RobotStatus struct {
	RobotID   RobotID                `json:"robot_id"`
	State     string                 `json:"state"`   // e.g. "IDLE", "BUSY", "ERROR", "CHARGING"
	Position  string                 `json:"position"`
	Battery   float64                `json:"battery"` // 0-100
	UpdatedAt time.Time              `json:"updated_at"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// ZoneSummary is aggregated zone state reported to area.
type ZoneSummary struct {
	ZoneID     ZoneID    `json:"zone_id"`
	RobotCount int       `json:"robot_count"`
	Healthy    int       `json:"healthy"`
	Busy       int       `json:"busy"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AreaSummary is aggregated area state reported to fleet.
type AreaSummary struct {
	AreaID     AreaID    `json:"area_id"`
	ZoneCount  int       `json:"zone_count"`
	RobotCount int       `json:"robot_count"`
	UpdatedAt  time.Time `json:"updated_at"`
}
