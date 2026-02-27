// Package api defines types for firmware updates and heterogeneous robot firmware.
package api

import "time"

// RobotCommandType for firmware and other extended commands.
const (
	RobotCommandTypeTask          = "TASK"
	RobotCommandTypeFirmwareUpdate = "FIRMWARE_UPDATE"
	RobotCommandTypeFirmwareRollback = "FIRMWARE_ROLLBACK"
)

// RobotStatus Extra keys for firmware (reported by edge in RobotStatus.Extra).
const (
	ExtraModelID             = "model_id"               // e.g. "picker-v2", "agv-x1"
	ExtraFirmwareVersion     = "firmware_version"      // e.g. "2.1.0"
	ExtraFirmwareUpdateStatus = "firmware_update_status" // idle | downloading | applying | success | failed | rollback
)

// FirmwareUpdateStatus values for ExtraFirmwareUpdateStatus.
const (
	FirmwareStatusIdle        = "idle"
	FirmwareStatusDownloading = "downloading"
	FirmwareStatusApplying     = "applying"
	FirmwareStatusSuccess      = "success"
	FirmwareStatusFailed       = "failed"
	FirmwareStatusRollback     = "rollback"
)

// FirmwareUpdatePayload is the JSON payload for RobotCommand type FIRMWARE_UPDATE.
// Edge uses this to download, verify, and apply without further calls to fleet/area/zone.
type FirmwareUpdatePayload struct {
	CampaignID      string `json:"campaign_id"`
	Version         string `json:"version"`
	ModelID         string `json:"model_id"`
	DownloadURL     string `json:"download_url"`
	ChecksumSHA256  string `json:"checksum_sha256"`
	RollbackVersion string `json:"rollback_version,omitempty"`
	RollbackURL     string `json:"rollback_url,omitempty"`
	Deadline        string `json:"deadline,omitempty"` // RFC3339
}

// FirmwareRollbackPayload is the JSON payload for RobotCommand type FIRMWARE_ROLLBACK.
type FirmwareRollbackPayload struct {
	CampaignID     string `json:"campaign_id"`
	Version        string `json:"version"` // version to apply (the rollback target)
	DownloadURL    string `json:"download_url"`
	ChecksumSHA256 string `json:"checksum_sha256"`
}

// FirmwareImage is a single firmware artifact in the catalog (fleet-side).
type FirmwareImage struct {
	ModelID         string    `json:"model_id"`
	Version         string    `json:"version"`
	DownloadURL     string    `json:"download_url"`
	ChecksumSHA256  string    `json:"checksum_sha256"`
	RollbackVersion string    `json:"rollback_version,omitempty"`
	RollbackURL     string    `json:"rollback_url,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// FirmwareCampaignTarget defines which robots get an update (fleet-side targeting).
type FirmwareCampaignTarget struct {
	ModelID           string   `json:"model_id"`
	TargetVersion     string   `json:"target_version"`
	CurrentVersion    string   `json:"current_version,omitempty"` // optional: only robots with this version
	ZoneIDs           []string `json:"zone_ids,omitempty"`       // optional: only these zones
	AreaIDs           []string `json:"area_ids,omitempty"`        // optional: only these areas
	MaxConcurrentPerZone int    `json:"max_concurrent_per_zone"`  // e.g. 50
	HealthGateSuccessRate float64 `json:"health_gate_success_rate"` // e.g. 0.99
	HealthGateMinCount int    `json:"health_gate_min_count"`     // min robots in stage to evaluate gate
}

// FirmwareCampaign is a fleet-level firmware rollout (staged by zone/area).
type FirmwareCampaign struct {
	ID        string                   `json:"id"`
	Target    FirmwareCampaignTarget    `json:"target"`
	Image     FirmwareImage             `json:"image"`
	Stages    []FirmwareCampaignStage   `json:"stages"` // order of rollout (e.g. zone-1, then zone-2, ...)
	Status    string                    `json:"status"` // pending | running | paused | completed | rolled_back
	CreatedAt time.Time                 `json:"created_at"`
}

// FirmwareCampaignStage is one stage of a rollout (e.g. one zone or one area).
type FirmwareCampaignStage struct {
	StageIndex int      `json:"stage_index"`
	ZoneIDs    []string `json:"zone_ids,omitempty"`
	AreaIDs    []string `json:"area_ids,omitempty"`
}
