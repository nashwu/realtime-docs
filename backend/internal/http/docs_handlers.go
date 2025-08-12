package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"realtime-docs/internal/store"
	"realtime-docs/pkg/auth"
)

type DocsAPI struct{ DB *store.Postgres }

type createDocReq struct {
	Title string `json:"title"`
}

type docResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Version   int64     `json:"version"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Create handles new doc creation for the authenticated user.
func (a *DocsAPI) Create(w http.ResponseWriter, r *http.Request) {
	var req createDocReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	uid := auth.UserID(r.Context())
	d, err := a.DB.CreateDoc(r.Context(), req.Title, uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(docResponse{
		ID: d.ID, Title: d.Title, Version: d.Version, UpdatedAt: d.UpdatedAt,
	})
}

// List returns up to 100 docs
func (a *DocsAPI) List(w http.ResponseWriter, r *http.Request) {
	docs, err := a.DB.ListDocs(r.Context(), 100, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := make([]docResponse, 0, len(docs))
	for _, d := range docs {
		resp = append(resp, docResponse{
			ID: d.ID, Title: d.Title, Version: d.Version, UpdatedAt: d.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// Get streams a doc's raw bytes and version header.
func (a *DocsAPI) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	d, err := a.DB.GetDoc(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("X-Doc-Version", fmt.Sprintf("%d", d.Version))
	_, _ = w.Write(d.Bytes)
}