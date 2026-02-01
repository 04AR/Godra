package ws

import (
	"context"
	"log"
	"sync"

	"github.com/example/godra/internal/gamestate"
)

type GameRoom struct {
	ID        string
	Clients   map[*Client]bool
	Broadcast chan []byte
	Cancel    context.CancelFunc
}

type Hub struct {
	clients map[*Client]bool
	register chan *Client
	unregister chan *Client
	rooms map[string]*GameRoom
	mu    sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]*GameRoom),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			room := h.getOrCreateRoom(client.GameID)
			room.Clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				if room, exists := h.rooms[client.GameID]; exists {
					delete(room.Clients, client)
					if len(room.Clients) == 0 {
						// Clean up room
						room.Cancel()
						delete(h.rooms, client.GameID)
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) getOrCreateRoom(gameID string) *GameRoom {
	if room, ok := h.rooms[gameID]; ok {
		return room
	}

	ctx, cancel := context.WithCancel(context.Background())
	room := &GameRoom{
		ID:        gameID,
		Clients:   make(map[*Client]bool),
		Broadcast: make(chan []byte),
		Cancel:    cancel,
	}
	h.rooms[gameID] = room

	// Start subscription listener for this room
	go room.listenToRedis(ctx)

	log.Printf("Created game room %s", gameID)
	return room
}

func (r *GameRoom) listenToRedis(ctx context.Context) {
	pubsub := gamestate.SubscribeToGame(ctx, r.ID)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}
			// Broadcast to all clients in this room
			for client := range r.Clients {
				select {
				case client.Send <- []byte(msg.Payload):
				default:
					close(client.Send)
					delete(r.Clients, client)
				}
			}
		}
	}
}
