package traceability

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"time"
)

//go:embed static/index.html
var dashboardHTML []byte

// Server is the Traceability HTTP API server.
type Server struct {
	Service *Service
}

// Handler returns the HTTP handler: dashboard at / and /ui, API elsewhere.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/records", s.handleRecords)
	mux.HandleFunc("/genealogy", s.handleGenealogy)
	mux.HandleFunc("/recall", s.handleRecall)
	mux.HandleFunc("/stats", s.handleStats)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "layer": "traceability"})
}

func (s *Server) handleRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var rec TraceRecord
	if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	created, err := s.Service.Record(r.Context(), &rec)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (s *Server) handleGenealogy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	serial := r.URL.Query().Get("serial")
	lot := r.URL.Query().Get("lot")
	if serial != "" {
		list := s.Service.GetGenealogyBySerial(r.Context(), serial)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"serial": serial, "records": list})
		return
	}
	if lot != "" {
		list := s.Service.GetGenealogyByLot(r.Context(), lot)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"lot": lot, "records": list})
		return
	}
	http.Error(w, "query parameter serial= or lot= required", http.StatusBadRequest)
}

func (s *Server) handleRecall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	lot := r.URL.Query().Get("lot")
	sku := r.URL.Query().Get("sku")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	var from, to *time.Time
	if fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err == nil {
			from = &t
		}
	}
	if toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err == nil {
			to = &t
		}
	}
	list := s.Service.Recall(r.Context(), lot, sku, from, to)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"records": list})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	n := s.Service.RecordCount(r.Context())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"total_records": n})
}
