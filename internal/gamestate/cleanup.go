package gamestate

import (
	"context"
	"godra/internal/metrics"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// StartSessionCleaner starts a background worker that cleans up expired guest sessions
func StartSessionCleaner(ctx context.Context, interval time.Duration, expirySeconds int) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cleanupExpiredSessions(expirySeconds)
			}
		}
	}()
}

func cleanupExpiredSessions(expirySeconds int) {
	// Calculate cutoff timestamp
	now := time.Now().Unix()
	cutoff := now - int64(expirySeconds)

	// Get expired users from Sorted Set
	// ZRANGEBYSCORE active_sessions:guests -inf <cutoff>
	users, err := RDB.ZRangeByScore(context.Background(), "active_sessions:guests", &redis.ZRangeBy{
		Min: "-inf",
		Max: strconv.FormatInt(cutoff, 10),
	}).Result()

	if err != nil {
		metrics.Log.Error("Failed to scan expired sessions", "error", err)
		return
	}

	for _, userID := range users {
		metrics.Log.Info("Cleaning up expired session", "user_id", userID)

		// Remove from Sorted Set
		RDB.ZRem(context.Background(), "active_sessions:guests", userID)

		// 2. Execute Disconnect Logic
		ExecuteScript(context.Background(), "on_disconnect", []string{}, userID)
	}
}
