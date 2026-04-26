package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/bubbletrack/server/internal/domain"
	"github.com/google/uuid"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// Client is a middleman between the websocket connection and the hub
type Client struct {
	ID   string
	Hub  *Hub
	Send chan []byte
}

// ChatMessage represents a chat message
type ChatMessage struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Sender    string `json:"sender"`
	Content   string `json:"content"`
	IsUser    bool   `json:"is_user"`
	CreatedAt string `json:"created_at"`
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					h.mu.RUnlock()
					h.mu.Lock()
					delete(h.clients, client)
					close(client.Send)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastMessage broadcasts a message to all connected clients
func (h *Hub) BroadcastMessage(msg ChatMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	h.broadcast <- data
	return nil
}

// BroadcastAnalysis broadcasts an analysis result to all clients
func (h *Hub) BroadcastAnalysis(result *domain.AnalysisResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	h.broadcast <- data
	return nil
}

// NewClient creates a new client
func NewClient(hub *Hub) *Client {
	return &Client{
		ID:   uuid.New().String(),
		Hub:  hub,
		Send: make(chan []byte, 256),
	}
}

func (h *Hub) NotifyChatMessage(userID, sender, content string, isUser bool) error {
	msg := ChatMessage{
		ID:        uuid.New().String(),
		UserID:    userID,
		Sender:    sender,
		Content:   content,
		IsUser:    isUser,
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	return h.BroadcastMessage(msg)
}
