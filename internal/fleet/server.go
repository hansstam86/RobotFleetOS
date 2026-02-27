package fleet

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/robotfleetos/robotfleetos/pkg/api"
)

//go:embed static/index.html
var dashboardHTML []byte

//go:embed static/maintenance.html
var maintenanceHTML []byte

// RecentWorkOrderEntry is a summary of a work order for the recent list.
type RecentWorkOrderEntry struct {
	ID            string `json:"id"`
	AreaID        string `json:"area_id"`
	Priority      int    `json:"priority"`
	PayloadSummary string `json:"payload_summary"` // e.g. "SKU SCOOTER-001 x 1000" or "firmware 2.0.0"
	CreatedAt     string `json:"created_at"`
}

// Server is the fleet HTTP API server.
type Server struct {
	Scheduler     *Scheduler
	State         *GlobalState
	recentMu      sync.RWMutex
	recentOrders  []RecentWorkOrderEntry
	maxRecent     int
}

// CreateWorkOrderRequest is the JSON body for POST /work_orders.
type CreateWorkOrderRequest struct {
	AreaID   string  `json:"area_id"`
	Priority int     `json:"priority"`
	Payload  string  `json:"payload"` // raw JSON or text; stored as UTF-8 bytes
	Deadline *string `json:"deadline,omitempty"` // RFC3339
}

// CreateWorkOrderResponse is the JSON response for POST /work_orders.
type CreateWorkOrderResponse struct {
	ID        string `json:"id"`
	AreaID    string `json:"area_id"`
	CreatedAt string `json:"created_at"`
}

// Handler returns an http.Handler that serves the dashboard at "/" and "/ui" and delegates API routes to the mux.
// This avoids relying on ServeMux's "/" pattern behavior, which can vary by Go version.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/work_orders", s.handleWorkOrders)
	mux.HandleFunc("/firmware/simulate", s.handleFirmwareSimulate)
	mux.HandleFunc("/state", s.handleGetState)
	mux.HandleFunc("/state/areas", s.handleGetAreas)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == "/" || p == "/ui" || p == "/ui/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write(dashboardHTML)
			return
		}
		if p == "/maintenance" || p == "/maintenance/" || p == "/ui/maintenance" || p == "/ui/maintenance/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write(maintenanceHTML)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

// RegisterRoutes mounts fleet API routes on mux (for backwards compatibility).
// Prefer using Handler() so the dashboard is served correctly at "/" and "/ui".
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/work_orders", s.handleWorkOrders)
	mux.HandleFunc("/state", s.handleGetState)
	mux.HandleFunc("/state/areas", s.handleGetAreas)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "layer": "fleet"})
}

func (s *Server) handleWorkOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.handleListWorkOrders(w, r)
		return
	}
	if r.Method == http.MethodPost {
		s.handleCreateWorkOrder(w, r)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) addRecent(order *api.WorkOrder) {
	summary := payloadSummary(order.Payload)
	entry := RecentWorkOrderEntry{
		ID:             string(order.ID),
		AreaID:         string(order.AreaID),
		Priority:       order.Priority,
		PayloadSummary: summary,
		CreatedAt:      order.CreatedAt.Format(time.RFC3339),
	}
	s.recentMu.Lock()
	defer s.recentMu.Unlock()
	if s.recentOrders == nil {
		s.recentOrders = make([]RecentWorkOrderEntry, 0, 100)
		s.maxRecent = 100
	}
	if s.maxRecent <= 0 {
		s.maxRecent = 100
	}
	s.recentOrders = append(s.recentOrders, entry)
	if len(s.recentOrders) > s.maxRecent {
		s.recentOrders = s.recentOrders[len(s.recentOrders)-s.maxRecent:]
	}
}

func payloadSummary(payload []byte) string {
	if len(payload) == 0 {
		return "—"
	}
	var m map[string]interface{}
	if err := json.Unmarshal(payload, &m); err != nil {
		if len(payload) > 80 {
			return string(payload[:80]) + "..."
		}
		return string(payload)
	}
	if t, _ := m["type"].(string); t == "firmware_update" {
		ver, _ := m["version"].(string)
		if ver != "" {
			return "firmware " + ver
		}
		return "firmware update"
	}
	// CMMS maintenance
	if t, _ := m["type"].(string); t == "maintenance" {
		ver, _ := m["target_firmware_version"].(string)
		if ver != "" {
			return "firmware " + ver
		}
		cid, _ := m["cmms_work_order_id"].(string)
		name, _ := m["equipment_name"].(string)
		if cid != "" && name != "" {
			return "maintenance " + cid + " (" + name + ")"
		}
		if cid != "" {
			return "maintenance " + cid
		}
		return "maintenance"
	}
	// WMS tasks: pick, putaway, move
	if t, _ := m["type"].(string); t == "pick" || t == "putaway" || t == "move" {
		sku, _ := m["sku"].(string)
		qty, _ := m["quantity"].(float64)
		from, _ := m["from_location_id"].(string)
		to, _ := m["to_location_id"].(string)
		if sku != "" && qty > 0 {
			return fmt.Sprintf("%s %s × %.0f (%s → %s)", t, sku, qty, from, to)
		}
		return t
	}
	if sku, ok := m["sku"].(string); ok {
		qty, _ := m["quantity"].(float64)
		if qty > 0 {
			return fmt.Sprintf("%s × %.0f", sku, qty)
		}
		return sku
	}
	return string(payload)
}

func (s *Server) handleListWorkOrders(w http.ResponseWriter, r *http.Request) {
	s.recentMu.RLock()
	list := make([]RecentWorkOrderEntry, len(s.recentOrders))
	copy(list, s.recentOrders)
	s.recentMu.RUnlock()
	// Newest first
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"work_orders": list})
}

func (s *Server) handleCreateWorkOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.AreaID == "" {
		http.Error(w, "area_id required", http.StatusBadRequest)
		return
	}
	order := &api.WorkOrder{
		AreaID:   api.AreaID(req.AreaID),
		Priority: req.Priority,
		Payload:  []byte(req.Payload),
	}
	if req.Deadline != nil {
		t, err := time.Parse(time.RFC3339, *req.Deadline)
		if err != nil {
			http.Error(w, "invalid deadline: "+err.Error(), http.StatusBadRequest)
			return
		}
		order.Deadline = &t
	}
	if err := s.Scheduler.SubmitWorkOrder(r.Context(), order); err != nil {
		http.Error(w, "submit failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	s.addRecent(order)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateWorkOrderResponse{
		ID:        string(order.ID),
		AreaID:    string(order.AreaID),
		CreatedAt: order.CreatedAt.Format(time.RFC3339),
	})
}

func (s *Server) handleFirmwareSimulate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		SeedBusy int `json:"seed_busy"` // optional: submit this many work orders first so that many robots are BUSY and will defer firmware until done
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	// Optionally seed work orders so some robots are BUSY and will defer firmware until they finish.
	if body.SeedBusy > 0 {
		for i := 0; i < body.SeedBusy; i++ {
			seedOrder := &api.WorkOrder{
				AreaID:   api.AreaID("area-1"),
				Priority: 2,
				Payload:  []byte(`{"task":"work","seed":true}`),
			}
			_ = s.Scheduler.SubmitWorkOrder(r.Context(), seedOrder)
		}
	}

	// Firmware campaign: zone will broadcast to all robots; idle robots update immediately, busy ones defer until work complete.
	payload := map[string]string{
		"type":             "firmware_update",
		"campaign_id":      "sim-" + time.Now().Format("20060102150405"),
		"version":          "2.0.0",
		"model_id":         "stub-model",
		"download_url":     "https://cdn.example/fw/stub-model/2.0.0.bin",
		"checksum_sha256":  "simulated",
		"rollback_version": "1.0.0",
	}
	payloadBytes, _ := json.Marshal(payload)
	order := &api.WorkOrder{
		AreaID:   api.AreaID("area-1"),
		Priority: 1,
		Payload:  payloadBytes,
	}
	if err := s.Scheduler.SubmitWorkOrder(r.Context(), order); err != nil {
		http.Error(w, "submit failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	s.addRecent(order)
	resp := map[string]interface{}{
		"ok":       true,
		"message":  "firmware campaign submitted; zone broadcasts to all robots. Busy robots defer update until work complete.",
		"order_id": string(order.ID),
		"target":   "area-1 -> zone -> all robots (stub-model 1.0.0 -> 2.0.0)",
	}
	if body.SeedBusy > 0 {
		resp["seed_busy"] = body.SeedBusy
		resp["message"] = fmt.Sprintf("Submitted %d seed work orders, then firmware campaign. Busy robots will defer firmware until task complete.", body.SeedBusy)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleGetState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	areas := s.State.GetAllAreas()
	total := s.State.TotalRobots()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"areas":        areas,
		"total_robots": total,
	})
}

func (s *Server) handleGetAreas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	areas := s.State.GetAllAreas()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"areas": areas})
}
