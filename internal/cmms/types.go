package cmms

import "time"

// EquipmentStatus is the operational status of an asset.
type EquipmentStatus string

const (
	EquipmentOperational       EquipmentStatus = "operational"
	EquipmentUnderMaintenance   EquipmentStatus = "under_maintenance"
	EquipmentOutOfService      EquipmentStatus = "out_of_service"
)

// EquipmentType categorizes the asset.
type EquipmentType string

const (
	EquipmentTypeRobot   EquipmentType = "robot"
	EquipmentTypeZone    EquipmentType = "zone"
	EquipmentTypeMachine EquipmentType = "machine"
	EquipmentTypeOther   EquipmentType = "other"
)

// Equipment is a maintainable asset (robot, zone, machine).
type Equipment struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Type      EquipmentType    `json:"type"`
	AreaID    string           `json:"area_id,omitempty"`
	ZoneID    string           `json:"zone_id,omitempty"`
	Status    EquipmentStatus  `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// MWOMaintenanceType is preventive, corrective, inspection, or firmware upgrade.
type MWOMaintenanceType string

const (
	MWOPreventive       MWOMaintenanceType = "preventive"
	MWOCorrective       MWOMaintenanceType = "corrective"
	MWOInspection       MWOMaintenanceType = "inspection"
	MWOFirmwareUpgrade  MWOMaintenanceType = "firmware_upgrade"
)

// MWOStatus is the status of a maintenance work order.
type MWOStatus string

const (
	MWOStatusOpen        MWOStatus = "open"
	MWOStatusInProgress  MWOStatus = "in_progress"
	MWOStatusCompleted   MWOStatus = "completed"
	MWOStatusCancelled   MWOStatus = "cancelled"
)

// MaintenanceWorkOrder is a maintenance work order for a piece of equipment.
type MaintenanceWorkOrder struct {
	ID               string             `json:"id"`
	EquipmentID      string             `json:"equipment_id"`
	Type             MWOMaintenanceType `json:"type"`
	Status           MWOStatus          `json:"status"`
	Priority         int                `json:"priority"` // 1–5, 5 highest
	DueDate          *time.Time         `json:"due_date,omitempty"`
	Description      string             `json:"description"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	CompletedAt           *time.Time         `json:"completed_at,omitempty"`
	FleetWorkOrderID       string             `json:"fleet_work_order_id,omitempty"` // set when submitted to Fleet
	TargetFirmwareVersion string             `json:"target_firmware_version,omitempty"` // for type firmware_upgrade
}
