package httpx

import (
	"log/slog"
	"net/http"
	"time"

	"realtime-docs/internal/app"
	"realtime-docs/internal/store"
	"realtime-docs/internal/ws"
	"realtime-docs/pkg/auth"
	"realtime-docs/pkg/metrics"
)

// NewRouter wires up all HTTP routes, middleware, and handlers
func NewRouter(cfg app.Config, logger *slog.Logger, hub *ws.Hub, db *store.Postgres) http.Handler {
	mw := NewMiddleware(cfg)
	api := &DocsAPI{DB: db}

	// Auth API
	j := auth.New(cfg.JWTSecret)
	authAPI := &AuthAPI{DB: db, JWT: j}

	mux := http.NewServeMux()

	// Health / readiness / metrics
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))
	mux.Handle("/readyz",  http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))
	mux.Handle("/metrics", metrics.Handler())

	// WebSocket endpoint
	mux.Handle("/ws", http.HandlerFunc(hub.ServeWS))

	// Auth endpoints
	mux.Handle("/api/auth/register", http.HandlerFunc(authAPI.Register))
	mux.Handle("/api/auth/login",    http.HandlerFunc(authAPI.Login))
	mux.Handle("/api/auth/me",       mw.Auth(http.HandlerFunc(authAPI.Me)))

	// Docs endpoints (JWT-protected)
	mux.Handle("/api/docs", mw.Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost { api.Create(w, r); return }
		if r.Method == http.MethodGet  { api.List(w, r);  return }
		http.NotFound(w, r)
	})))
	mux.Handle("/api/docs/{id}", mw.Auth(http.HandlerFunc(api.Get)))

	// Server wrapper with read timeout
	s := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mw.Wrap(mux), // CORS + rate limit applied globally
		ReadHeaderTimeout: 10 * time.Second,
	}
	return s.Handler
}
