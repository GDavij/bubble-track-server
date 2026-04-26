package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type ChatAPIClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewChatAPIClient(baseURL string) *ChatAPIClient {
	return &ChatAPIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type sendMessageRequest struct {
	Sender    string `json:"sender"`
	Content   string `json:"content"`
	SessionID string `json:"session_id"`
}

type sendMessageResponse struct {
	Message struct {
		ID        string `json:"id"`
		UserID    string `json:"user_id"`
		Sender    string `json:"sender"`
		Content   string `json:"content"`
		IsUser    bool   `json:"is_user"`
		CreatedAt string `json:"created_at"`
	} `json:"message"`
}

type chatMessageDTO struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Sender    string `json:"sender"`
	Content   string `json:"content"`
	IsUser    bool   `json:"is_user"`
	CreatedAt string `json:"created_at"`
}

type getMessagesResponse struct {
	Messages []chatMessageDTO `json:"messages"`
}

func (c *ChatAPIClient) SendMessage(sender, content, sessionID string) error {
	reqBody := sendMessageRequest{
		Sender:    sender,
		Content:   content,
		SessionID: sessionID,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/chat", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var out sendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func (c *ChatAPIClient) GetMessages(sessionID string, limit int) ([]ChatMessage, error) {
	u, err := url.Parse(c.baseURL + "/api/chat")
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	q := u.Query()
	if sessionID != "" {
		q.Set("session_id", sessionID)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	u.RawQuery = q.Encode()

	resp, err := c.httpClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("get request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var out getMessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	msgs := make([]ChatMessage, len(out.Messages))
	for i, m := range out.Messages {
		timeLabel := m.CreatedAt
		if ts, parseErr := time.Parse(time.RFC3339Nano, m.CreatedAt); parseErr == nil {
			timeLabel = ts.Local().Format("15:04")
		}
		msgs[i] = ChatMessage{
			Sender:    m.Sender,
			Content:   m.Content,
			Timestamp: timeLabel,
			IsUser:    m.IsUser,
		}
	}

	return msgs, nil
}
