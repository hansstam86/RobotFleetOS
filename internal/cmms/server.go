package cmms

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

//go:embed static/index.html
var dashboardHTML []byte

// Server is the CMMS HTTP API server.
type Server struct {
	Service *Service
}

// Handler returns the HTTP handler.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/firmware/trigger", s.handleFirmwareTrigger)
	mux.HandleFunc("/equipment", s.handleEquipment)
	mux.HandleFunc("/equipment/", s.handleEquipmentByID)
	mux.HandleFunc("/mwo", s.handleMWO)
	mux.HandleFunc("/mwo/", s.handleMWOByID)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "layer": "cmms"})
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
	msg, err := s.Service.TriggerFirmwareCampaign(r.Context(), body.SeedBusy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": msg})
}

func (s *Server) handleEquipment(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		status := EquipmentStatus(r.URL.Query().Get("status"))
		eqType := EquipmentType(r.URL.Query().Get("type"))
		list := s.Service.ListEquipment(r.Context(), status, eqType)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"equipment": list})
		return
	case http.MethodPost:
		var e Equipment
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		created, err := s.Service.CreateEquipment(r.Context(), &e)
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

func (s *Server) handleEquipmentByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/equipment/")
	if id == "" {
		http.Error(w, "equipment id required", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		e := s.Service.GetEquipment(r.Context(), id)
		if e == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(e)
		return
	case http.MethodPut:
		var e Equipment
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		e.ID = id
		updated, err := s.Service.UpdateEquipment(r.Context(), &e)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleMWO(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		status := MWOStatus(r.URL.Query().Get("status"))
		equipmentID := r.URL.Query().Get("equipment_id")
		list := s.Service.ListMWOs(r.Context(), status, equipmentID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"work_orders": list})
		return
	case http.MethodPost:
		var m MaintenanceWorkOrder
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		created, err := s.Service.CreateMWO(r.Context(), &m)
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

func (s *Server) handleMWOByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/mwo/")
	parts := strings.SplitN(path, "/", 2)
	id := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}
	if id == "" {
		http.Error(w, "work order id required", http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodGet && action == "" {
		m := s.Service.GetMWO(r.Context(), id)
		if m == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
		return
	}
	if r.Method == http.MethodPost && action != "" {
		var result *MaintenanceWorkOrder
		var err error
		switch action {
		case "start":
			result, err = s.Service.StartMWO(r.Context(), id)
		case "complete":
			result, err = s.Service.CompleteMWO(r.Context(), id)
		case "cancel":
			result, err = s.Service.CancelMWO(r.Context(), id)
		case "submit_to_fleet":
			priority := 3
			if pr := r.URL.Query().Get("priority"); pr != "" {
				if p, ok := parseInt(pr); ok {
					priority = p
				}
			}
			fid, errF := s.Service.SubmitMWOToFleet(r.Context(), id, priority)
			if errF != nil {
				http.Error(w, errF.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"fleet_work_order_id": fid})
			return
		default:
			http.Error(w, "unknown action: "+action, http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func parseInt(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	return n, err == nil
}
