package mes

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
)

//go:embed static/index.html
var dashboardHTML []byte

// Server is the MES HTTP API server.
type Server struct {
	Service *Service
}

// Handler returns the HTTP handler: dashboard at / and /ui, API on /health, /orders, etc.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/firmware/trigger", s.handleFirmwareTrigger)
	mux.HandleFunc("/orders", s.handleOrders)
	mux.HandleFunc("/orders/", s.handleOrderByID)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == "/" || p == "/ui" || p == "/ui/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write(dashboardHTML)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "layer": "mes"})
}

func (s *Server) handleFirmwareTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		SeedBusy int `json:"seed_busy"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	msg, err := s.Service.TriggerFirmwareUpdate(r.Context(), body.SeedBusy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "message": msg})
}

// CreateOrderRequest is the JSON body for POST /orders.
type CreateOrderRequest struct {
	ERPOrderRef     string `json:"erp_order_ref,omitempty"`
	SKU             string `json:"sku"`
	ProductID        string `json:"product_id,omitempty"`
	Quantity         int    `json:"quantity"`
	AreaID           string `json:"area_id"`
	ZoneID           string `json:"zone_id,omitempty"`
	Priority         int    `json:"priority,omitempty"`
	BOMRevision      string `json:"bom_revision,omitempty"`
	RoutingRevision  string `json:"routing_revision,omitempty"`
}

func (s *Server) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleCreateOrder(w, r)
		return
	case http.MethodGet:
		s.handleListOrders(w, r)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	order := &ProductionOrder{
		ERPOrderRef:      req.ERPOrderRef,
		SKU:              req.SKU,
		ProductID:         req.ProductID,
		Quantity:         req.Quantity,
		AreaID:           req.AreaID,
		ZoneID:           req.ZoneID,
		Priority:         req.Priority,
		BOMRevision:      req.BOMRevision,
		RoutingRevision:  req.RoutingRevision,
	}
	if order.Priority <= 0 {
		order.Priority = 1
	}
	created, err := s.Service.CreateOrder(r.Context(), order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (s *Server) handleListOrders(w http.ResponseWriter, r *http.Request) {
	statusFilter := OrderStatus(r.URL.Query().Get("status"))
	list := s.Service.ListOrders(r.Context(), statusFilter)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"orders": list})
}

func (s *Server) handleOrderByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/orders/")
	if id == "" {
		http.Error(w, "order id required", http.StatusBadRequest)
		return
	}
	// Remove any trailing path (e.g. /orders/ord-1/release -> id = "ord-1/release", we need to split)
	if idx := strings.Index(id, "/"); idx >= 0 {
		action := id[idx+1:]
		id = id[:idx]
		switch action {
		case "release":
			s.handleReleaseOrder(w, r, id)
			return
		case "pause":
			s.handlePauseOrder(w, r, id)
			return
		case "complete":
			s.handleCompleteOrder(w, r, id)
			return
		case "cancel":
			s.handleCancelOrder(w, r, id)
			return
		case "scrap":
			s.handleReportScrap(w, r, id)
			return
		}
	}
	if r.Method == http.MethodGet {
		s.handleGetOrder(w, r, id)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleGetOrder(w http.ResponseWriter, r *http.Request, id string) {
	order := s.Service.GetOrder(r.Context(), id)
	if order == nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (s *Server) handleReleaseOrder(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	order, err := s.Service.ReleaseOrder(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (s *Server) handlePauseOrder(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	order, err := s.Service.PauseOrder(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (s *Server) handleCompleteOrder(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	order, err := s.Service.CompleteOrder(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (s *Server) handleCancelOrder(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	order, err := s.Service.CancelOrder(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// ScrapRequest is the JSON body for POST /orders/:id/scrap.
type ScrapRequest struct {
	Quantity int    `json:"quantity"`
	Reason   string `json:"reason"`
}

func (s *Server) handleReportScrap(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ScrapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	order, err := s.Service.ReportScrap(r.Context(), id, req.Quantity, req.Reason)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
