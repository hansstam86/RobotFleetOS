package erp

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
)

//go:embed static/index.html
var dashboardHTML []byte

type Server struct {
	Service *Service
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "layer": "erp"})
}

func (s *Server) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		status := OrderStatus(r.URL.Query().Get("status"))
		list := s.Service.ListOrders(r.Context(), status)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"orders": list})
		return
	case http.MethodPost:
		var o Order
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		created, err := s.Service.CreateOrder(r.Context(), &o)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleOrderByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/orders/")
	parts := strings.SplitN(path, "/", 2)
	id := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}
	if id == "" {
		http.Error(w, "order id required", http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodGet && action == "" {
		o := s.Service.GetOrder(r.Context(), id)
		if o == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(o)
		return
	}
	if r.Method == http.MethodPost && action != "" {
		switch action {
		case "submit_to_mes":
			var body struct {
				ZoneID   string `json:"zone_id"`
				Priority int    `json:"priority"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.Priority <= 0 {
				body.Priority = 1
			}
			o, err := s.Service.SubmitToMES(r.Context(), id, body.ZoneID, body.Priority)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(o)
			return
		case "cancel":
			o, err := s.Service.CancelOrder(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(o)
			return
		}
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
