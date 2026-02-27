package qms

import "time"

// InspectionResult is pass or fail.
type InspectionResult string

const (
	InspectionPass InspectionResult = "pass"
	InspectionFail InspectionResult = "fail"
)

// Inspection is a single inspection record (serial/lot at a station).
type Inspection struct {
	ID         string           `json:"id"`
	Serial     string           `json:"serial,omitempty"`
	Lot        string           `json:"lot,omitempty"`
	SKU        string           `json:"sku"`
	StationID  string           `json:"station_id,omitempty"`
	Result     InspectionResult `json:"result"`
	Notes      string           `json:"notes,omitempty"`
	MESOrderID string           `json:"mes_order_id,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
}

// NCRStatus is the status of a non-conformance report.
type NCRStatus string

const (
	NCRStatusOpen   NCRStatus = "open"
	NCRStatusClosed NCRStatus = "closed"
)

// NCR is a non-conformance report.
type NCR struct {
	ID          string    `json:"id"`
	Serial      string    `json:"serial,omitempty"`
	Lot         string    `json:"lot,omitempty"`
	SKU         string    `json:"sku"`
	Description string    `json:"description"`
	Status      NCRStatus `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	ClosedAt    *time.Time `json:"closed_at,omitempty"`
}

// Hold is a hold on a serial or lot (release when resolved).
type Hold struct {
	ID         string     `json:"id"`
	Serial     string     `json:"serial,omitempty"`
	Lot        string     `json:"lot,omitempty"`
	Reason     string     `json:"reason"`
	HeldAt     time.Time  `json:"held_at"`
	ReleasedAt *time.Time `json:"released_at,omitempty"`
}
