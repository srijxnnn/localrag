package chunk

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultSize    = 500
	DefaultOverlap = 100
)

// FromFile reads a .txt or .md file and returns overlapping text chunks.
func FromFile(path string) ([]string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".txt" && ext != ".md" {
		return nil, fmt.Errorf("unsupported file type %q (only .txt and .md)", ext)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Split(string(data), DefaultSize, DefaultOverlap), nil
}

// Split divides text into chunks of at most size runes, advancing by (size - overlap)
// runes between chunk starts so boundaries overlap by overlap runes.
func Split(text string, size, overlap int) []string {
	if size <= 0 || overlap < 0 || overlap >= size {
		return nil
	}
	runes := []rune(text)
	if len(runes) == 0 {
		return nil
	}
	step := size - overlap
	var chunks []string
	for start := 0; start < len(runes); start += step {
		end := min(start+size, len(runes))
		chunks = append(chunks, string(runes[start:end]))
		if end == len(runes) {
			break
		}
	}
	return chunks
}
