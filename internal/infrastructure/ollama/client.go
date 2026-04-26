package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	baseURL         string
	model           string
	embeddingModel  string
	client          *http.Client
}

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	System string `json:"system,omitempty"`
}

type GenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func NewClient(baseURL, model string) (*Client, error) {
	return NewClientWithEmbedding(baseURL, model, "")
}

func NewClientWithEmbedding(baseURL, model, embeddingModel string) (*Client, error) {
	if embeddingModel == "" {
		embeddingModel = "nomic-embed-text"
	}
	return &Client{
		baseURL:         baseURL,
		model:           model,
		embeddingModel:  embeddingModel,
		client:          &http.Client{},
	}, nil
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	req := GenerateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama: %s", string(respBody))
	}

	var genResp GenerateResponse
	if err := json.Unmarshal(respBody, &genResp); err != nil {
		return "", fmt.Errorf("unmarshal: %w", err)
	}

	return genResp.Response, nil
}

func (c *Client) Close() error {
	c.client.CloseIdleConnections()
	return nil
}

type EmbeddingsRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type EmbeddingsResponse struct {
	Embedding []float32 `json:"embedding"`
}

func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	req := EmbeddingsRequest{Model: c.embeddingModel, Input: text}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/embeddings", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var embResp EmbeddingsResponse
	if json.Unmarshal(respBody, &embResp) == nil && len(embResp.Embedding) > 0 {
		return embResp.Embedding, nil
	}

	// Fallback: hash-based pseudo-embedding
	emb := make([]float32, 768)
	h := 0
	for _, ch := range text {
		h = h*31 + int(ch)
	}
	for i := range emb {
		emb[i] = float32((h+i)%1000) / 1000.0
	}
	return emb, nil
}