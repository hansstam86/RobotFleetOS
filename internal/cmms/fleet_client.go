package cmms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// FleetClient submits maintenance work orders to the Fleet layer.
type FleetClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewFleetClient returns a client for the given Fleet API base URL.
func NewFleetClient(baseURL string) *FleetClient {
	return &FleetClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// SubmitMaintenanceWorkOrder sends a maintenance task to Fleet. areaID is used as work order area_id.
// payload can include cmms_work_order_id, equipment_id, description. Returns the Fleet work order ID.
func (c *FleetClient) SubmitMaintenanceWorkOrder(ctx context.Context, areaID string, priority int, payload map[string]interface{}) (fleetWorkOrderID string, err error) {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	reqBody := map[string]interface{}{
		"area_id":  areaID,
		"priority": priority,
		"payload":  string(bodyBytes),
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/work_orders", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("fleet API returned %d", resp.StatusCode)
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.ID, nil
}

// TriggerFirmwareSimulate triggers a firmware campaign on Fleet (POST /firmware/simulate).
// seedBusy is the number of robots to mark busy so they defer the update; 0 for none.
func (c *FleetClient) TriggerFirmwareSimulate(ctx context.Context, seedBusy int) (message string, err error) {
	reqBody := map[string]interface{}{}
	if seedBusy > 0 {
		reqBody["seed_busy"] = seedBusy
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/firmware/simulate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("fleet firmware API returned %d: %s", resp.StatusCode, out.Error)
	}
	return out.Message, nil
}
