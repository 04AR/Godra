package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DBType       string // "sqlite" or "postgres"
	DBDSN        string
	RedisAddr    string
	SyncInterval int
}

func Load() *Config {
	// Load .env (ignore error if file missing)
	_ = godotenv.Load()

	cfg := &Config{}

	// Define defaults using Env Vars if present
	defaultPort := getEnv("PORT", "8080")
	defaultDBType := getEnv("DB_TYPE", "sqlite")
	defaultDBDSN := getEnv("DB_DSN", "game.db")
	defaultRedisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	defaultSyncInterval, _ := strconv.Atoi(getEnv("SYNC_INTERVAL", "50"))

	// Parse Flags (override defaults/env)
	flag.StringVar(&cfg.Port, "port", defaultPort, "Server port")
	flag.StringVar(&cfg.DBType, "db-type", defaultDBType, "Database type: sqlite or postgres")
	flag.StringVar(&cfg.DBDSN, "db-dsn", defaultDBDSN, "Database DSN (e.g., game.db or postgres://user:pass@localhost:5432/dbname)")
	flag.StringVar(&cfg.RedisAddr, "redis-addr", defaultRedisAddr, "Redis address")
	flag.IntVar(&cfg.SyncInterval, "sync-interval", defaultSyncInterval, "Sync interval in milliseconds")

	flag.Parse()

	return cfg
}

// Helper to get env var with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
