package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client holds the base URL of the local Ollama server.
type Client struct {
	BaseURL string
}

func New(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

// Embed sends text to Ollama and gets back a vector (slice of float64).
func (c *Client) Embed(model, text string) ([]float64, error) {
	body, _ := json.Marshal(map[string]string{
		"model":  model,
		"prompt": text,
	})

	resp, err := http.Post(c.BaseURL+"/api/embeddings", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("Ollama is not running. Start it with: ollama serve")
	}
	defer resp.Body.Close()

	var result struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ollama embed decode: %w", err)
	}
	if len(result.Embedding) == 0 {
		return nil, fmt.Errorf("model %q is not pulled. Run: ollama pull %s", model, model)
	}
	return result.Embedding, nil
}

// Generate sends a prompt to Ollama and returns the model's response as a string.
func (c *Client) Generate(model, prompt string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	})

	resp, err := http.Post(c.BaseURL+"/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama generate: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ollama generate decode: %w", err)
	}
	return result.Response, nil
}
