package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// Upgrader handles upgrading HTTP connections to WebSocket
type Upgrader struct {
	hub      *Hub
	upgrader websocket.Upgrader
}

// NewUpgrader creates a new WebSocket upgrader
func NewUpgrader(hub *Hub) *Upgrader {
	return &Upgrader{
		hub: hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// RegisterRoutes registers WebSocket routes
func (u *Upgrader) RegisterRoutes(e *echo.Echo) {
	e.GET("/ws", u.handleWebSocket)
}

func (u *Upgrader) handleWebSocket(c echo.Context) error {
	conn, err := u.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	client := NewClient(u.hub)
	u.hub.register <- client

	defer func() {
		u.hub.unregister <- client
		conn.Close()
	}()

	go u.writePump(c.Request().Context(), client, conn)
	u.readPump(c.Request().Context(), client, conn)

	return nil
}

func (u *Upgrader) readPump(ctx context.Context, client *Client, conn *websocket.Conn) {
	conn.SetReadLimit(4096)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}

			if userID, ok := msg["user_id"].(string); ok {
				client.ID = userID
			}
		}
	}
}

func (u *Upgrader) writePump(ctx context.Context, client *Client, conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		case message, ok := <-client.Send:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
