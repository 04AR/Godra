package api

import (
	"encoding/json"
	"net/http"

	"github.com/example/godra/internal/auth"
	"github.com/example/godra/internal/gamestate"
)

type RPCRequest struct {
	Script string        `json:"script"`
	Args   []interface{} `json:"args"`
	Keys   []string      `json:"keys"`
}

type RPCResponse struct {
	Result interface{} `json:"result"`
}

func RPCHandler(w http.ResponseWriter, r *http.Request) {
	// Auth first
	tokenString := r.Header.Get("Authorization")
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}
	
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse requests
	var req RPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Script == "" {
		http.Error(w, "Script name required", http.StatusBadRequest)
		return
	}

	// Verify Permissions
	requiredRole := gamestate.GetScriptRole(req.Script)
	// Role Hierarchy: manager > player
	// Simple check: if required is manager, user must be manager.
	// If required is player, anyone (who is authenticated) access.
	// If it's guest, anyone.
	if requiredRole == "manager" && claims.Role != "manager" {
		http.Error(w, "Forbidden: Manager role required", http.StatusForbidden)
		return
	}

	// Inject UserContext
	// We prepend User ID to args to ensure scripts always know who is calling
	// Convention: ARGV[1] is always User ID.
	userID := claims.UserID
	finalArgs := append([]interface{}{userID}, req.Args...)

	// Execute
	// We now support custom KEYS via RPC if provided.
	result, err := gamestate.ExecuteScript(r.Context(), req.Script, req.Keys, finalArgs...)
	if err != nil {
		http.Error(w, "Script execution failed: " + err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RPCResponse{Result: result})
}
