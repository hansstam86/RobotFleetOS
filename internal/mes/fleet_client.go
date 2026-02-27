package mes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// FleetClient submits work orders to the Fleet layer HTTP API.
type FleetClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewFleetClient returns a client for the given Fleet API base URL (e.g. "http://localhost:8080").
func NewFleetClient(baseURL string) *FleetClient {
	return &FleetClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// SubmitWorkOrderRequest is the payload sent to Fleet POST /work_orders.
type SubmitWorkOrderRequest struct {
	AreaID   string  `json:"area_id"`
	Priority int     `json:"priority"`
	Payload  string  `json:"payload"`
	Deadline *string `json:"deadline,omitempty"`
}

// SubmitWorkOrderResponse is the response from Fleet POST /work_orders.
type SubmitWorkOrderResponse struct {
	ID        string `json:"id"`
	AreaID    string `json:"area_id"`
	CreatedAt string `json:"created_at"`
}

// SubmitWorkOrder sends a work order to Fleet. Returns the Fleet work order ID.
func (c *FleetClient) SubmitWorkOrder(ctx context.Context, areaID string, priority int, payload []byte, deadline *time.Time) (fleetWorkOrderID string, err error) {
	payloadStr := string(payload)
	if payloadStr == "" {
		payloadStr = "{}"
	}
	reqBody := SubmitWorkOrderRequest{
		AreaID:   areaID,
		Priority: priority,
		Payload:  payloadStr,
	}
	if deadline != nil {
		s := deadline.Format(time.RFC3339)
		reqBody.Deadline = &s
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
	var out SubmitWorkOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.ID, nil
}

// FirmwareSimulate triggers a firmware update simulation on Fleet (POST /firmware/simulate).
// seedBusy is the number of work orders to submit first so that many robots are BUSY and defer the update.
func (c *FleetClient) FirmwareSimulate(ctx context.Context, seedBusy int) (message string, err error) {
	body := map[string]int{"seed_busy": seedBusy}
	if seedBusy <= 0 {
		body = nil
	}
	var bodyReader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader([]byte("{}"))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/firmware/simulate", bodyReader)
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
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("fleet API returned %d", resp.StatusCode)
	}
	return out.Message, nil
}
