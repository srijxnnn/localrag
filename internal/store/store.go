package store

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"math"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const (
	// RAGDir is the workspace directory created by `rag init`.
	RAGDir = ".rag"
	// DBName is the SQLite file inside RAGDir.
	DBName = "data.db"
)

// DBPath returns the path to the default database file (relative to cwd).
func DBPath() string {
	return filepath.Join(RAGDir, DBName)
}

// Store wraps the SQLite database for documents and embedding chunks.
type Store struct {
	db *sql.DB
}

// Chunk is one indexed text segment with its embedding vector.
type Chunk struct {
	ID           int64
	DocumentID   int64
	DocumentPath string
	Text         string
	Embedding    []float64
}

// ListedDocument is one indexed file with metadata for `rag list`.
type ListedDocument struct {
	ID         int64
	Path       string
	AddedAt    int64 // Unix seconds
	ChunkCount int64
}

// Init opens or creates the database at dbPath, applies schema, and returns a Store.
func Init(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("store init pragma: %w", err)
	}

	steps := []string{
		`CREATE TABLE IF NOT EXISTS documents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL UNIQUE,
			added_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS chunks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			document_id INTEGER NOT NULL,
			text TEXT NOT NULL,
			embedding BLOB NOT NULL,
			FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
		)`,
	}
	for _, q := range steps {
		if _, err := db.Exec(q); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("store init schema: %w", err)
		}
	}

	return &Store{db: db}, nil
}

// Close releases the database handle.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// SaveDocument inserts a document row and returns its id.
func (s *Store) SaveDocument(path string) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO documents (path, added_at) VALUES (?, ?)`,
		path, time.Now().Unix(),
	)
	if err != nil {
		return 0, fmt.Errorf("save document: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ListDocuments returns every document with its chunk count, ordered by path.
func (s *Store) ListDocuments() ([]ListedDocument, error) {
	rows, err := s.db.Query(`
		SELECT d.id, d.path, d.added_at, COUNT(c.id)
		FROM documents d
		LEFT JOIN chunks c ON c.document_id = d.id
		GROUP BY d.id
		ORDER BY d.path ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	defer rows.Close()

	var out []ListedDocument
	for rows.Next() {
		var d ListedDocument
		if err := rows.Scan(&d.ID, &d.Path, &d.AddedAt, &d.ChunkCount); err != nil {
			return nil, fmt.Errorf("list documents scan: %w", err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list documents rows: %w", err)
	}
	return out, nil
}

// SaveChunk stores one chunk and its embedding vector for a document.
func (s *Store) SaveChunk(docID int64, text string, embedding []float64) error {
	blob := EncodeEmbedding(embedding)
	_, err := s.db.Exec(
		`INSERT INTO chunks (document_id, text, embedding) VALUES (?, ?, ?)`,
		docID, text, blob,
	)
	if err != nil {
		return fmt.Errorf("save chunk: %w", err)
	}
	return nil
}

// EncodeEmbedding packs float64s as little-endian bytes for BLOB storage.
func EncodeEmbedding(v []float64) []byte {
	buf := make([]byte, 8*len(v))
	for i, f := range v {
		binary.LittleEndian.PutUint64(buf[i*8:], math.Float64bits(f))
	}
	return buf
}

// DecodeEmbedding unpacks a BLOB from EncodeEmbedding back to float64s.
func DecodeEmbedding(blob []byte) ([]float64, error) {
	if len(blob)%8 != 0 {
		return nil, fmt.Errorf("invalid embedding blob length %d", len(blob))
	}
	n := len(blob) / 8
	out := make([]float64, n)
	for i := range out {
		u := binary.LittleEndian.Uint64(blob[i*8:])
		out[i] = math.Float64frombits(u)
	}
	return out, nil
}

// AllChunks loads every chunk with its document path and decoded embedding.
func (s *Store) AllChunks() ([]Chunk, error) {
	rows, err := s.db.Query(`
		SELECT c.id, c.document_id, d.path, c.text, c.embedding
		FROM chunks c
		JOIN documents d ON d.id = c.document_id
	`)
	if err != nil {
		return nil, fmt.Errorf("all chunks query: %w", err)
	}
	defer rows.Close()

	var out []Chunk
	for rows.Next() {
		var (
			c          Chunk
			embeddingB []byte
		)
		if err := rows.Scan(&c.ID, &c.DocumentID, &c.DocumentPath, &c.Text, &embeddingB); err != nil {
			return nil, fmt.Errorf("all chunks scan: %w", err)
		}
		emb, err := DecodeEmbedding(embeddingB)
		if err != nil {
			return nil, err
		}
		c.Embedding = emb
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("all chunks rows: %w", err)
	}
	return out, nil
}
