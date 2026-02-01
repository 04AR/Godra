package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
    
	"github.com/example/godra/internal/api"
	"github.com/example/godra/internal/auth"
	"github.com/example/godra/internal/config"
	"github.com/example/godra/internal/database"
	"github.com/example/godra/internal/gamestate"
	"github.com/example/godra/internal/logger"
	"github.com/example/godra/internal/metrics"
	"github.com/example/godra/internal/ws"
)

func main() {
	// 0. Init Logger
	logger.Init()

	// 1. Load Config
	cfg := config.Load()
	log.Printf("Starting Godra Server on port %s...", cfg.Port)

	// 2. Init Database
	if err := database.Init(cfg.DBType, cfg.DBDSN); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	// 3. Init Redis
	if err := gamestate.Init(cfg.RedisAddr); err != nil {
		log.Fatalf("Redis initialization failed: %v", err)
	}
	// Start Cleanup Worker
	gamestate.StartSessionCleaner(context.Background(), 5*time.Second, 10)

	// 4. Init WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// 5. Setup Router
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(api.RequestLogger) // [NEW] Our structured logger

	r.Post("/register", auth.RegisterHandler)
	r.Post("/login", auth.LoginHandler)
	r.Post("/guest-login", auth.GuestLoginHandler)
	r.Post("/api/rpc", api.RPCHandler) // Generic Endpoint
	
	r.Get("/metrics", metrics.Handler) // [NEW] Metrics Endpoint

	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	// 6. Start Server
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
