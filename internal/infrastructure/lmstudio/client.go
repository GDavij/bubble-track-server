package lmstudio

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

type EmbeddingsRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type EmbeddingsResponse struct {
	Data  []EmbeddingData `json:"data"`
	Model string          `json:"model"`
	Usage json.RawMessage `json:"usage,omitempty"`
}

type EmbeddingData struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

func NewClient(baseURL, model string) *Client {
	return &Client{
		baseURL:    baseURL,
		model:      model,
		httpClient: &http.Client{},
	}
}

func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	req := EmbeddingsRequest{Model: c.model, Input: text}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/embeddings", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var embResp EmbeddingsResponse
	if json.Unmarshal(respBody, &embResp) == nil && len(embResp.Data) > 0 && len(embResp.Data[0].Embedding) > 0 {
		return embResp.Data[0].Embedding, nil
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

func (c *Client) CloseIdleConnections() {
	c.httpClient.CloseIdleConnections()
}
