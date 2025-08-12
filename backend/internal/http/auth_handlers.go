package httpx

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"realtime-docs/internal/store"
	"realtime-docs/pkg/auth"
)

type AuthAPI struct {
	DB  *store.Postgres
	JWT *auth.JWT
}

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type tokenResp struct {
	Token string      `json:"token"`
	User  authUserDTO `json:"user"`
}
type authUserDTO struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// Register handles user signup and returns a JWT
func (a *AuthAPI) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad payload", http.StatusBadRequest)
		return
	}
	req.Email = strings.TrimSpace(req.Email)

	// Basic validation
	if len(req.Password) < 8 || !strings.Contains(req.Email, "@") {
		http.Error(w, "invalid email or weak password", http.StatusBadRequest)
		return
	}

	// Create user
	u, err := a.DB.CreateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "email already in use", http.StatusConflict)
		return
	}

	// Issue token for 24hrs
	tok, _ := a.JWT.Sign(u.ID, 24*time.Hour)
	writeJSON(w, tokenResp{Token: tok, User: authUserDTO{ID: u.ID, Email: u.Email}})
}

// Login verifies credentials and returns a JWT
func (a *AuthAPI) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad payload", http.StatusBadRequest)
		return
	}

	// Check credentials
	u, err := a.DB.VerifyUser(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Issue token (24h)
	tok, _ := a.JWT.Sign(u.ID, 24*time.Hour)
	writeJSON(w, tokenResp{Token: tok, User: authUserDTO{ID: u.ID, Email: u.Email}})
}

// Me returns the authenticated user's ID
func (a *AuthAPI) Me(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	if uid == "anon" || uid == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeJSON(w, map[string]string{"userId": uid})
}

// send JSON with proper headers
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
