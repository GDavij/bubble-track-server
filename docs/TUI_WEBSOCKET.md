# TUI WebSocket Integration

## Overview

The Bubble Track TUI now supports optional WebSocket connectivity for real-time chat message updates. When a WebSocket connection is available, the TUI will receive messages instantly via WebSocket broadcasts. If WebSocket is not available or fails to connect, the TUI automatically falls back to HTTP polling.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    TUI (Terminal UI)                          │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  WebSocket Client (optional)                        │   │
│  │  - Connects to server on WS_URL                     │   │
│  │  - Receives real-time message broadcasts             │   │
│  │  - Auto-reconnects on connection loss               │   │
│  └─────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  HTTP Client (fallback)                             │   │
│  │  - POST /api/chat to send messages                  │   │
│  │  - GET /api/chat to retrieve message history        │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                          ↕
┌─────────────────────────────────────────────────────────────┐
│                    Bubble Track Server                         │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  WebSocket Hub                                       │   │
│  │  - Broadcasts chat messages to all connected clients│   │
│  │  - Manages client connections                        │   │
│  └─────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  API Handlers                                        │   │
│  │  - POST /api/chat saves to DB                       │   │
│  │  - Broadcasts to WebSocket hub after save             │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Configuration

Add the following environment variables to your configuration:

```bash
# WebSocket URL (optional - if not set, TUI uses HTTP only)
WS_URL=ws://localhost:8080/ws

# API URL (required - used for HTTP send/receive)
API_URL=http://localhost:8080
```

## Usage

### Starting the Server with WebSocket Support

The server automatically starts WebSocket support when the hub is initialized:

```bash
# Start server (WebSocket hub starts automatically)
go run ./cmd/server

# The WebSocket endpoint is available at: ws://localhost:8080/ws
```

### Starting the TUI with WebSocket Client

```bash
# Set WebSocket URL (optional)
export WS_URL=ws://localhost:8080/ws
export API_URL=http://localhost:8080

# Start TUI
go run ./cmd/tui
```

The TUI will:
1. Attempt to connect to WebSocket endpoint
2. On success, display "Connected (WS)" status
3. Receive real-time message broadcasts
4. If connection fails, fall back to HTTP
5. Display "Disconnected" or "WS Error" status if WebSocket fails

### Connection Status Indicators

The TUI chat panel shows the current connection status:

- **"Ready"** - HTTP-only mode (no WebSocket configured)
- **"Connected (WS)"** - WebSocket connected, receiving real-time updates
- **"Live"** - actively receiving messages via WebSocket
- **"Disconnected"** - WebSocket connection lost
- **"WS Error"** - WebSocket connection failed
- **"Sent"** - Message sent successfully
- **"History loaded (N)"** - Historical messages loaded

## WebSocket Client Features

### Connection Management

- **Automatic reconnection**: Client attempts to reconnect with exponential backoff
- **Retry logic**: Configurable max retries (default: 3)
- **Connection monitoring**: Status displayed in TUI interface
- **Graceful degradation**: HTTP fallback if WebSocket unavailable

### Message Flow

1. User sends message via TUI
2. TUI POSTs to `/api/chat` (HTTP)
3. Server saves to database
4. Server broadcasts to WebSocket hub
5. All connected TUI clients receive message via WebSocket
6. TUI displays message in real-time

### Code Structure

#### WebSocket Client (internal/tui/websocket/chat.go)

```go
client := websocket.NewClient(wsURL, maxRetries, reconnect)
err := client.Connect(ctx)
go client.Run(ctx)

// Receive messages
for data := range client.Receive() {
    var msg websocket.ChatMessage
    json.Unmarshal(data, &msg)
    // Process message
}

// Cleanup
client.Close()
```

#### TUI Model Integration (internal/tui/model.go)

The TUI model accepts an optional WebSocket client:

```go
model := tui.NewModelWithChatService(
    graphRepo,
    chatClient,
    "test-user",
    logger,
    wsClient, // Optional - nil = HTTP only
)
```

#### Server Handler (internal/api/handler.go)

The handler broadcasts to WebSocket hub after saving messages:

```go
if h.wsHub != nil {
    wsMsg := websocket.ChatMessage{
        ID:        msg.ID.String(),
        UserID:    msg.UserID,
        Sender:    msg.Sender,
        Content:   msg.Content,
        IsUser:    msg.IsUser,
        CreatedAt: msg.CreatedAt.Format(time.RFC3339Nano),
    }
    h.wsHub.BroadcastMessage(wsMsg)
}
```

## Message Format

WebSocket messages use the following JSON format:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "test-user",
  "sender": "You",
  "content": "Hello, world!",
  "is_user": true,
  "created_at": "2026-04-26T10:30:00.123Z"
}
```

## Troubleshooting

### WebSocket Connection Fails

1. Check `WS_URL` environment variable is set correctly
2. Verify server is running and `/ws` endpoint is accessible
3. Check firewall/network restrictions
4. TUI will automatically fall back to HTTP

### Messages Not Real-Time

1. Verify `WS_URL` is set
2. Check TUI status shows "Connected (WS)" not "Disconnected"
3. Ensure server WebSocket hub is running
4. Check server logs for WebSocket broadcast errors

### Connection Drops Intermittently

1. Network connectivity issues
2. WebSocket ping timeout (54 seconds)
3. Server restarts
4. Client automatically reconnects with backoff

## Performance Considerations

- **WebSocket**: Real-time, low latency, persistent connection
- **HTTP**: Polling-based, higher latency, request/response cycle
- **Fallback**: HTTP fallback ensures messages always delivered even if WebSocket fails

## Security Notes

- WebSocket endpoint (`/ws`) currently has no authentication
- Consider adding authentication middleware for production use
- Messages are broadcast to all connected clients
- User-specific filtering should be added for multi-user environments

## Testing

Build and test the executables:

```bash
# Build TUI
go build -o bubble-tui ./cmd/tui

# Build server
go build -o bubble-server ./cmd/server

# Run tests
go test ./...
```

## Next Steps

- [ ] Add WebSocket authentication
- [ ] Implement message filtering by user/session
- [ ] Add connection statistics and monitoring
- [ ] Implement WebSocket compression
- [ ] Add client-specific message routing
