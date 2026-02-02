package metrics

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

var (
	TotalRequests     atomic.Uint64
	ActiveConnections atomic.Int64
	ActiveLobbies     atomic.Int64
)

func Handler(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"total_requests":     TotalRequests.Load(),
		"active_connections": ActiveConnections.Load(),
		"active_lobbies":     ActiveLobbies.Load(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
