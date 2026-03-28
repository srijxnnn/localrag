package search

import (
	"math"
	"sort"

	"github.com/srijxnnn/localrag/internal/store"
)

// Cosine returns cosine similarity between a and b in [-1, 1], or 0 if
// lengths differ, either slice is empty, or either vector has zero magnitude.
func Cosine(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot float64
	for i := range a {
		dot += a[i] * b[i]
	}
	ma, mb := magnitude(a), magnitude(b)
	if ma == 0 || mb == 0 {
		return 0
	}
	return dot / (ma * mb)
}

func magnitude(v []float64) float64 {
	var s float64
	for _, x := range v {
		s += x * x
	}
	return math.Sqrt(s)
}

// TopK returns up to k chunks with highest cosine similarity to the query embedding.
func TopK(query []float64, chunks []store.Chunk, k int) []store.Chunk {
	if k <= 0 || len(chunks) == 0 {
		return nil
	}
	type scored struct {
		c store.Chunk
		s float64
	}
	scores := make([]scored, len(chunks))
	for i := range chunks {
		scores[i] = scored{c: chunks[i], s: Cosine(query, chunks[i].Embedding)}
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].s > scores[j].s })
	k = min(k, len(scores))
	out := make([]store.Chunk, k)
	for i := 0; i < k; i++ {
		out[i] = scores[i].c
	}
	return out
}
