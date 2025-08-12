package httpx

import (
	"net/http"
	"strings"
	"time"

	"github.com/rs/cors"
	"realtime-docs/internal/app"
	"realtime-docs/pkg/auth"
	"realtime-docs/pkg/ratelimit"
)

type Middleware struct {
	cors   *cors.Cors
	auth   *auth.JWT
	rlimit *ratelimit.Limiter
}

// NewMiddleware builds the shared middleware stack from config
func NewMiddleware(cfg app.Config) *Middleware {
	return &Middleware{
		cors: cors.New(cors.Options{
			AllowedOrigins:   cfg.CORSAllow,
			AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		}),
		auth:   auth.New(cfg.JWTSecret),
		rlimit: ratelimit.New(30, time.Minute), // 30 req/min default
	}
}

// Wrap applies CORS + rate limiting to a handler
func (m *Middleware) Wrap(h http.Handler) http.Handler {
	return m.cors.Handler(m.rlimit.Middleware(h))
}

// Auth enforces JWT auth and adds user ID to the request context
func (m *Middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := r.Header.Get("Authorization")
		if !strings.HasPrefix(b, "Bearer ") {
			http.Error(w, "no token", http.StatusUnauthorized)
			return
		}
		tok := strings.TrimPrefix(b, "Bearer ")
		uid, err := m.auth.Verify(tok)
		if err != nil {
			http.Error(w, "bad token", http.StatusUnauthorized)
			return
		}
		// Pass along the user ID for downstream handlers
		next.ServeHTTP(w, r.WithContext(auth.WithUser(r.Context(), uid)))
	})
}