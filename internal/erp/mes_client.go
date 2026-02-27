package erp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MESClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewMESClient(baseURL string) *MESClient {
	return &MESClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// CreateProductionOrder creates a production order in MES. Returns MES order ID.
func (c *MESClient) CreateProductionOrder(ctx context.Context, erpOrderRef, sku string, quantity int, areaID, zoneID string, priority int) (mesOrderID string, err error) {
	body := map[string]interface{}{
		"erp_order_ref": erpOrderRef,
		"sku":           sku,
		"quantity":      quantity,
		"area_id":       areaID,
		"priority":      priority,
	}
	if zoneID != "" {
		body["zone_id"] = zoneID
	}
	if priority <= 0 {
		body["priority"] = 1
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/orders", bytes.NewReader(raw))
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
		return "", fmt.Errorf("mes API returned %d", resp.StatusCode)
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.ID, nil
}
