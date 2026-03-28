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
