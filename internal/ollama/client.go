package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	if resp.StatusCode != http.StatusOK {
		return "", errGenerateHTTP(resp, model)
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ollama generate decode: %w", err)
	}
	return result.Response, nil
}

// GenerateStream streams the generated text to out (one POST, NDJSON lines with stream: true).
func (c *Client) GenerateStream(model, prompt string, out io.Writer) error {
	body, err := json.Marshal(map[string]any{
		"model":  model,
		"prompt": prompt,
		"stream": true,
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(c.BaseURL+"/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("ollama generate: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errGenerateHTTP(resp, model)
	}

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var chunk struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}
		if err := json.Unmarshal(line, &chunk); err != nil {
			continue
		}
		if chunk.Response != "" {
			if _, err := io.WriteString(out, chunk.Response); err != nil {
				return err
			}
		}
		if chunk.Done {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ollama generate stream: %w", err)
	}
	return nil
}

func errGenerateHTTP(resp *http.Response, model string) error {
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("model %q is not available (pull it or pass --model). Run: ollama pull %s", model, model)
	}
	return fmt.Errorf("ollama generate: %s", resp.Status)
}
