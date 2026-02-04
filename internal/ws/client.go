package ws

import (
	"context"
	"log"
	"net/http"
	"time"

	"godra/internal/auth"
	"godra/internal/gamestate"
	"godra/internal/metrics"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	UserID   string
	Username string
	GameID   string
}

type IncomingMessage struct {
	Action string `json:"action"`
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// 1. Auth check
	token := r.URL.Query().Get("token")
	claims, err := auth.ValidateToken(token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	gameID := r.URL.Query().Get("game_id")
	if gameID == "" {
		http.Error(w, "Missing game_id", http.StatusBadRequest)
		return
	}

	// We execute "on_connect" script which validates lobby and joins user
	gameKey := "game:" + gameID
	playersKey := gameKey + ":players"
	_, err = gamestate.ExecuteScript(r.Context(), "on_connect", []string{gameKey, playersKey}, claims.Username, gameID)
	if err != nil {
		log.Printf("Connection rejected by on_connect hook: %v", err)
		http.Error(w, "Connection rejected: "+err.Error(), http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	metrics.ActiveConnections.Add(1)

	client := &Client{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   claims.UserID,
		Username: claims.Username,
		GameID:   gameID,
	}

	client.Hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	// Cleanup Guest
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
		metrics.ActiveConnections.Add(^int64(0))

		if len(c.UserID) > 6 && c.UserID[:6] == "guest:" {
			// Clean up guest data via Lua
			gamestate.ExecuteScript(context.Background(), "on_disconnect", []string{}, c.UserID)
		}
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Process message execute Lua script
		log.Printf("Player %s sent action: %s", c.Username, string(message))

		gameKey := "game:" + c.GameID
		_, err = gamestate.ExecuteScript(context.Background(), "update_state", []string{gameKey}, c.Username, string(message))
		if err != nil {
			log.Printf("Error updating state: %v", err)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(50 * time.Millisecond) // hardcode 50ms default
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	// Buffer for batching
	var buffer [][]byte

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// Buffer message
			buffer = append(buffer, message)

		case <-ticker.C:
			if len(buffer) > 0 {
				if len(buffer) == 1 {
					c.Conn.WriteMessage(websocket.TextMessage, buffer[0])
				} else {
					// Combine into batch JSON
					w, err := c.Conn.NextWriter(websocket.TextMessage)
					if err != nil {
						return
					}

					w.Write([]byte(`{"type":"batch","events":[`))
					for i, msg := range buffer {
						if i > 0 {
							w.Write([]byte(`,`))
						}
						w.Write(msg)
					}
					w.Write([]byte(`]}`))

					if err := w.Close(); err != nil {
						return
					}
				}
				buffer = nil
			}
		}
	}
}
