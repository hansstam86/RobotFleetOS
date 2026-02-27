package qms

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
)

//go:embed static/index.html
var dashboardHTML []byte

// Server is the QMS HTTP API server.
type Server struct {
	Service *Service
}

// Handler returns the HTTP handler.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/inspections", s.handleInspections)
	mux.HandleFunc("/ncr", s.handleNCR)
	mux.HandleFunc("/ncr/", s.handleNCRByID)
	mux.HandleFunc("/holds", s.handleHolds)
	mux.HandleFunc("/holds/", s.handleHoldByID)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "layer": "qms"})
}

func (s *Server) handleInspections(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var rec Inspection
		if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		created, err := s.Service.RecordInspection(r.Context(), &rec)
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
		serial := r.URL.Query().Get("serial")
		lot := r.URL.Query().Get("lot")
		list := s.Service.ListInspections(r.Context(), serial, lot)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"inspections": list})
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleNCR(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var n NCR
		if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		created, err := s.Service.CreateNCR(r.Context(), &n)
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
		status := NCRStatus(r.URL.Query().Get("status"))
		list := s.Service.ListNCRs(r.Context(), status)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ncrs": list})
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleNCRByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/ncr/")
	if idx := strings.Index(id, "/"); idx >= 0 {
		action := id[idx+1:]
		id = id[:idx]
		if action == "close" && r.Method == http.MethodPost {
			n, err := s.Service.CloseNCR(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(n)
			return
		}
	}
	if r.Method == http.MethodGet {
		n := s.Service.GetNCR(r.Context(), id)
		if n == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(n)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleHolds(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var h Hold
		if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		created, err := s.Service.CreateHold(r.Context(), &h)
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
		activeOnly := r.URL.Query().Get("active") == "true" || r.URL.Query().Get("active") == "1"
		list := s.Service.ListHolds(r.Context(), activeOnly)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"holds": list})
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleHoldByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/holds/")
	if idx := strings.Index(id, "/"); idx >= 0 {
		action := id[idx+1:]
		id = id[:idx]
		if action == "release" && r.Method == http.MethodPost {
			h, err := s.Service.ReleaseHold(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(h)
			return
		}
	}
	if r.Method == http.MethodGet {
		h := s.Service.GetHold(r.Context(), id)
		if h == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(h)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
