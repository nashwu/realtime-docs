package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	app "realtime-docs/internal/app"
	httpx "realtime-docs/internal/http"
	store "realtime-docs/internal/store"
	ws "realtime-docs/internal/ws"
)

func main() {
	// Load local .env (dev only)
	_ = godotenv.Load()

	cfg := app.LoadConfig()
	logger := app.NewLogger(cfg.Env)

	// Cancel on SIGINT/SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Postgres connection + migrations
	pg, err := store.NewPostgres(ctx, cfg, logger)
	if err != nil {
		logger.Error("postgres connect", "err", err)
		log.Fatal(err)
	}
	defer pg.Close()
	if err := store.RunMigrations(ctx, pg, logger); err != nil {
		logger.Error("migrations", "err", err)
		log.Fatal(err)
	}

	// Redis bus for WS fanout
	bus, err := ws.NewRedisBus(ctx, cfg, logger)
	if err != nil {
		logger.Error("redis connect", "err", err)
		log.Fatal(err)
	}
	defer bus.Close()

	// WebSocket hub
	hub := ws.NewHub(logger, bus, pg)
	go hub.Run(ctx)

	// HTTP + WS router
	router := httpx.NewRouter(cfg, logger, hub, pg)
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start server
	go func() {
		logger.Info("server.listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server.crash", "err", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	logger.Info("server.shutdown.start")

	// shutdown
	shutdownCtx, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()
	_ = srv.Shutdown(shutdownCtx)

	logger.Info("server.shutdown.complete")
	_ = os.Stdout.Sync()
}
