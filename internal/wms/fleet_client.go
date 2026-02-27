package wms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// FleetClient submits warehouse work orders to the Fleet layer HTTP API.
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

// SubmitWorkOrder sends a work order to Fleet. Returns the Fleet work order ID.
func (c *FleetClient) SubmitWorkOrder(ctx context.Context, areaID string, priority int, payload []byte) (fleetWorkOrderID string, err error) {
	payloadStr := string(payload)
	if payloadStr == "" {
		payloadStr = "{}"
	}
	reqBody := map[string]interface{}{
		"area_id":  areaID,
		"priority": priority,
		"payload":  payloadStr,
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
