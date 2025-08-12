package app

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	Env       string
	HTTPAddr  string
	CORSAllow []string

	JWTSecret string

	PGURL     string // e.g. postgres://user:pass@localhost:5432/docs?sslmode=disable
	PGMaxConn int

	RedisAddr string // host:port
	RedisDB   int
}

func LoadConfig() Config {
	cfg := Config{
		Env:       getEnv("APP_ENV", "dev"),
		HTTPAddr:  getEnv("HTTP_ADDR", ":8080"),
		JWTSecret: getEnv("JWT_SECRET", "dev-secret-change"),
		PGURL:     getEnv("PG_URL", "postgres://postgres:secret@localhost:5432/docs?sslmode=disable"),
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
	}
	cfg.PGMaxConn = getEnvInt("PG_MAX_CONN", 10)
	cfg.RedisDB = getEnvInt("REDIS_DB", 0)
	// CORS allowlist
	allow := getEnv("CORS_ALLOW", "http://localhost:4200")
	cfg.CORSAllow = splitCSV(allow)
	log.Printf("config: %+v\n", cfg)
	return cfg
}


// getEnv returns the env var or a default
func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// getEnvInt parses an int env var with a fallback
func getEnvInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		var i int
		_, _ = fmt.Sscanf(v, "%d", &i)
		if i > 0 {
			return i
		}
	}
	return def
}

// splitCSV trims and filters a comma-separated list
func splitCSV(v string) []string {
	var out []string
	for _, s := range strings.Split(v, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}