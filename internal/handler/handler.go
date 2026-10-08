package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"consulta-resultado-votacao-presidente-2026-oz-gemini/internal/data"
	"consulta-resultado-votacao-presidente-2026-oz-gemini/internal/model"
)

type Handler struct {
	repo *data.Repository
}

func NewHandler(repo *data.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/summary", h.handleSummary)
	mux.HandleFunc("/api/candidates", h.handleCandidates)
	mux.HandleFunc("/api/zones", h.handleZones)
	mux.HandleFunc("/api/search", h.handleSearch)
	mux.HandleFunc("/api/section", h.handleSection)

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) handleSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	summary, err := h.repo.GetSummary()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, summary)
}

func (h *Handler) handleCandidates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	candidates, err := h.repo.GetCandidateTotals()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, candidates)
}

func (h *Handler) handleZones(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	zones, err := h.repo.GetZones()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, zones)
}

func (h *Handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()

	zona, _ := strconv.Atoi(q.Get("zona"))
	secao, _ := strconv.Atoi(q.Get("secao"))
	candidate := strings.TrimSpace(q.Get("candidate"))
	local := strings.TrimSpace(q.Get("local"))

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 100
	}
	offset, _ := strconv.Atoi(q.Get("offset"))

	results, err := h.repo.SearchVotes(data.Filter{
		Zona:      zona,
		Secao:     secao,
		Candidate: candidate,
		Local:     local,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if results == nil {
		results = []model.SectionResult{} // return empty slice if nil
	}
	respondJSON(w, http.StatusOK, results)
}

func (h *Handler) handleSection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	zona, _ := strconv.Atoi(q.Get("zona"))
	secao, _ := strconv.Atoi(q.Get("secao"))

	if zona <= 0 || secao <= 0 {
		http.Error(w, "zona and secao parameters are required and must be positive integers", http.StatusBadRequest)
		return
	}

	detail, err := h.repo.GetSectionDetail(zona, secao)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if detail == nil {
		http.Error(w, "secao not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, detail)
}
