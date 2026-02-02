package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/example/godra/internal/database"
	"github.com/example/godra/internal/gamestate"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type AuthResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	UserID   uint   `json:"user_id"`
	Role     string `json:"role"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	role := "player"
	if req.Role == "manager" {
		role = "manager"
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	user := database.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     role,
	}

	if result := database.DB.Create(&user); result.Error != nil {
		http.Error(w, "Error creating user (username might be taken)", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user database.User
	if result := database.DB.Where("username = ?", req.Username).First(&user); result.Error != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Compare Hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	userID := fmt.Sprintf("%d", user.ID)
	token, err := GenerateToken(userID, user.Username, user.Role)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	resp := AuthResponse{
		Token:    token,
		Username: user.Username,
		UserID:   user.ID,
		Role:     user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func GuestLoginHandler(w http.ResponseWriter, r *http.Request) {
	// Simple Guest Login
	// Generate random Guest ID
	guestID := fmt.Sprintf("guest:%s", database.GenerateRandomString(8))
	
    // Store in Redis (Temporary)
    // We treat "guest:xyz" as a key with dummy value
    gamestate.RDB.Set(r.Context(), guestID, "active", 24*time.Hour)

	token, err := GenerateToken(guestID, "Guest", "guest")
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}
	
	resp := map[string]string{
		"token": token,
		"user_id": guestID,
		"role": "guest",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
