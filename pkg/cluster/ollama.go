package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	DefaultOllamaURL = "http://localhost:11434"
	EmbeddingModel   = "nomic-embed-text"
	EmbeddingTimeout = 30 * time.Second
)

type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
	available  bool
}

type embeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

func NewOllamaClient() *OllamaClient {
	client := &OllamaClient{
		baseURL: DefaultOllamaURL,
		httpClient: &http.Client{
			Timeout: EmbeddingTimeout,
		},
	}
	client.available = client.checkAvailability()
	return client
}

func NewOllamaClientWithURL(url string) *OllamaClient {
	client := &OllamaClient{
		baseURL: url,
		httpClient: &http.Client{
			Timeout: EmbeddingTimeout,
		},
	}
	client.available = client.checkAvailability()
	return client
}

func (c *OllamaClient) checkAvailability() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func (c *OllamaClient) IsAvailable() bool {
	return c.available
}

func (c *OllamaClient) Refresh() bool {
	c.available = c.checkAvailability()
	return c.available
}

func (c *OllamaClient) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	if !c.available {
		return nil, fmt.Errorf("ollama not available")
	}

	reqBody := embeddingRequest{
		Model:  EmbeddingModel,
		Prompt: text,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/embeddings", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var embResp embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return embResp.Embedding, nil
}

func (c *OllamaClient) GetEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	if !c.available {
		return nil, fmt.Errorf("ollama not available")
	}

	embeddings := make([][]float64, len(texts))

	for i, text := range texts {
		emb, err := c.GetEmbedding(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("getting embedding %d: %w", i, err)
		}
		embeddings[i] = emb
	}

	return embeddings, nil
}

func CheckOllama() bool {
	client := NewOllamaClient()
	return client.IsAvailable()
}
