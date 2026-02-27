package plm

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
	mux.HandleFunc("/products", s.handleProducts)
	mux.HandleFunc("/products/", s.handleProductByID)
	mux.HandleFunc("/ecos", s.handleECOs)
	mux.HandleFunc("/ecos/", s.handleECOByID)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "layer": "plm"})
}

func (s *Server) handleProducts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		status := ProductStatus(r.URL.Query().Get("status"))
		sku := r.URL.Query().Get("sku")
		list := s.Service.ListProducts(r.Context(), status, sku)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"products": list})
		return
	case http.MethodPost:
		var p Product
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		created, err := s.Service.CreateProduct(r.Context(), &p)
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

func (s *Server) handleProductByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/products/")
	parts := strings.SplitN(path, "/", 3)
	id := parts[0]
	sub := ""
	if len(parts) > 1 {
		sub = parts[1]
	}
	if id == "" {
		http.Error(w, "product id required", http.StatusBadRequest)
		return
	}
	if sub == "bom" {
		s.handleBOM(w, r, id, parts)
		return
	}
	switch r.Method {
	case http.MethodGet:
		p := s.Service.GetProduct(r.Context(), id)
		if p == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
		return
	case http.MethodPut:
		var p Product
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		p.ID = id
		updated, err := s.Service.UpdateProduct(r.Context(), &p)
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

func (s *Server) handleBOM(w http.ResponseWriter, r *http.Request, productID string, parts []string) {
	switch r.Method {
	case http.MethodGet:
		lines := s.Service.GetBOM(r.Context(), productID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"bom": lines})
		return
	case http.MethodPost:
		var b BOMLine
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		created, err := s.Service.AddBOMLine(r.Context(), productID, &b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
		return
	case http.MethodDelete:
		if len(parts) < 3 {
			http.Error(w, "bom line id required", http.StatusBadRequest)
			return
		}
		lineID := parts[2]
		if err := s.Service.DeleteBOMLine(r.Context(), lineID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleECOs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		status := ECOStatus(r.URL.Query().Get("status"))
		productID := r.URL.Query().Get("product_id")
		list := s.Service.ListECOs(r.Context(), status, productID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ecos": list})
		return
	case http.MethodPost:
		var e ECO
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		created, err := s.Service.CreateECO(r.Context(), &e)
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

func (s *Server) handleECOByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/ecos/")
	if id == "" {
		http.Error(w, "eco id required", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		e := s.Service.GetECO(r.Context(), id)
		if e == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(e)
		return
	case http.MethodPut:
		var e ECO
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		e.ID = id
		updated, err := s.Service.UpdateECO(r.Context(), &e)
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
