package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	url        string
	conn       *websocket.Conn
	send       chan []byte
	receive    chan []byte
	done       chan struct{}
	log        *slog.Logger
	reconnect  bool
	maxRetries int
}

type Message struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
}

type ChatMessage struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Sender    string `json:"sender"`
	Content   string `json:"content"`
	IsUser    bool   `json:"is_user"`
	CreatedAt string `json:"created_at"`
}

func NewClient(url string, maxRetries int, reconnect bool) *Client {
	return &Client{
		url:        url,
		send:       make(chan []byte, 256),
		receive:    make(chan []byte, 256),
		done:       make(chan struct{}),
		log:        slog.Default(),
		reconnect:  reconnect,
		maxRetries: maxRetries,
	}
}

func (c *Client) Connect(ctx context.Context) error {
	var lastErr error
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		c.log.Info("connecting to websocket", "url", c.url, "attempt", attempt+1)

		dialer := websocket.DefaultDialer
		conn, _, err := dialer.DialContext(ctx, c.url, nil)
		if err == nil {
			c.conn = conn
			c.log.Info("websocket connected", "url", c.url)
			return nil
		}

		lastErr = err
		c.log.Warn("websocket connection failed, retrying", "attempt", attempt+1, "error", err)

		backoff := time.Duration(attempt+1) * time.Second
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return fmt.Errorf("failed to connect after %d attempts: %w", c.maxRetries, lastErr)
}

func (c *Client) Run(ctx context.Context) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	go c.readPump(ctx)
	go c.writePump(ctx)
	<-ctx.Done()
	c.Close()
	return nil
}

func (c *Client) readPump(ctx context.Context) {
	defer c.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				c.log.Error("websocket read error", "error", err)
				return
			}

			select {
			case c.receive <- message:
			default:
				c.log.Warn("receive channel full, dropping message")
			}
		}
	}
}

func (c *Client) writePump(ctx context.Context) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.log.Error("websocket write error", "error", err)
				return
			}

		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.log.Error("websocket ping error", "error", err)
				return
			}
		}
	}
}

func (c *Client) Send(message interface{}) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
		return nil
	default:
		return fmt.Errorf("send channel full")
	}
}

func (c *Client) Receive() <-chan []byte {
	return c.receive
}

func (c *Client) Close() {
	select {
	case <-c.done:
		return
	default:
		close(c.done)
	}

	if c.conn != nil {
		c.conn.Close()
	}

	close(c.send)
	close(c.receive)
}

func (c *Client) IsConnected() bool {
	return c.conn != nil
}

func (c *Client) GetURL() string {
	return c.url
}
