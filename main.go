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
	"github.com/example/godra/internal/database"
	"github.com/example/godra/internal/gamestate"
	"github.com/example/godra/internal/metrics"
	"github.com/example/godra/internal/ws"
)

func main() {
	// Init Logger
	metrics.Init()

	// Load Config
	cfg := Load()
	log.Printf("Starting Godra Server on port %s...", cfg.Port)

	// Init DB
	if err := database.Init(cfg.DBType, cfg.DBDSN); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	// Init Redis
	if err := gamestate.Init(cfg.RedisAddr); err != nil {
		log.Fatalf("Redis initialization failed: %v", err)
	}
	// Start Cleanup Worker
	gamestate.StartSessionCleaner(context.Background(), 5*time.Second, 10)

	// Init WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// Setup Router
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(api.RequestLogger)

	r.Post("/register", auth.RegisterHandler)
	r.Post("/login", auth.LoginHandler)
	r.Post("/guest-login", auth.GuestLoginHandler)
	r.Post("/api/rpc", api.RPCHandler)
	
	r.Get("/metrics", metrics.Handler)

	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	//Start Server
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}


