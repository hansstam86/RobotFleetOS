package wms

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
)

//go:embed static/index.html
var dashboardHTML []byte

// Server is the WMS HTTP API server.
type Server struct {
	Service *Service
}

// Handler returns the HTTP handler: dashboard at / and /ui, API elsewhere.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/locations", s.handleLocations)
	mux.HandleFunc("/inventory", s.handleInventory)
	mux.HandleFunc("/tasks", s.handleTasks)
	mux.HandleFunc("/tasks/", s.handleTaskByID)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "layer": "wms"})
}

func (s *Server) handleLocations(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var loc Location
		if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.Service.CreateLocation(r.Context(), &loc); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(loc)
		return
	}
	if r.Method == http.MethodGet {
		list := s.Service.ListLocations(r.Context())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"locations": list})
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req struct {
			LocationID string `json:"location_id"`
			SKU        string `json:"sku"`
			Quantity   int    `json:"quantity"`
			Lot        string `json:"lot,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.Service.ReceiveInventory(r.Context(), req.LocationID, req.SKU, req.Quantity, req.Lot); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"ok": "received"})
		return
	}
	if r.Method == http.MethodGet {
		locationID := r.URL.Query().Get("location_id")
		sku := r.URL.Query().Get("sku")
		list := s.Service.ListInventory(r.Context(), locationID, sku)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"inventory": list})
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var t Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		created, err := s.Service.CreateTask(r.Context(), &t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
		return
	}
	if r.Method == http.MethodGet {
		status := TaskStatus(r.URL.Query().Get("status"))
		taskType := TaskType(r.URL.Query().Get("type"))
		list := s.Service.ListTasks(r.Context(), status, taskType)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"tasks": list})
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if idx := strings.Index(id, "/"); idx >= 0 {
		action := id[idx+1:]
		id = id[:idx]
		switch action {
		case "release":
			s.handleTaskRelease(w, r, id)
			return
		case "complete":
			s.handleTaskComplete(w, r, id)
			return
		case "cancel":
			s.handleTaskCancel(w, r, id)
			return
		}
	}
	if r.Method == http.MethodGet {
		t := s.Service.GetTask(r.Context(), id)
		if t == nil {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleTaskRelease(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	t, err := s.Service.ReleaseTask(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (s *Server) handleTaskComplete(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	t, err := s.Service.CompleteTask(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (s *Server) handleTaskCancel(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	t, err := s.Service.CancelTask(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}
