//go:build cgo

package cache

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteCache struct {
	db *sql.DB
}

func NewSQLiteCache(path string) (*SQLiteCache, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_cache_size=10000")
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("sqlite ping: %w", err)
	}

	if err := execAll(db, []string{
		`CREATE TABLE IF NOT EXISTS embeddings_cache (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			model_id TEXT NOT NULL,
			input_hash TEXT NOT NULL,
			embedding BLOB NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(model_id, input_hash)
		)`,
		`CREATE TABLE IF NOT EXISTS response_cache (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			model_id TEXT NOT NULL,
			input_hash TEXT NOT NULL,
			response TEXT NOT NULL,
			parameters TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(model_id, input_hash, parameters)
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			model_id TEXT NOT NULL,
			messages TEXT NOT NULL,
			context BLOB,
			metadata TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_embeddings_lookup ON embeddings_cache(model_id, input_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_response_lookup ON response_cache(model_id, input_hash, parameters)`,
	}); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite init: %w", err)
	}

	return &SQLiteCache{db: db}, nil
}

func (c *SQLiteCache) GetEmbedding(ctx context.Context, modelID, input string) ([]float32, error) {
	hash := hashInput(input)

	var data []byte
	err := c.db.QueryRowContext(ctx,
		`SELECT embedding FROM embeddings_cache WHERE model_id = ? AND input_hash = ?`,
		modelID, hash,
	).Scan(&data)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return bytesToFloats(data), nil
}

func (c *SQLiteCache) SetEmbedding(ctx context.Context, modelID, input string, embedding []float32) error {
	hash := hashInput(input)
	data := floatsToBytes(embedding)

	_, err := c.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO embeddings_cache (model_id, input_hash, embedding) VALUES (?, ?, ?)`,
		modelID, hash, data,
	)
	return err
}

func (c *SQLiteCache) GetResponse(ctx context.Context, modelID, input, params string) (string, error) {
	hash := hashInput(input)

	var response string
	err := c.db.QueryRowContext(ctx,
		`SELECT response FROM response_cache WHERE model_id = ? AND input_hash = ? AND parameters = ?`,
		modelID, hash, params,
	).Scan(&response)

	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return response, nil
}

func (c *SQLiteCache) SetResponse(ctx context.Context, modelID, input, params, response string) error {
	hash := hashInput(input)

	_, err := c.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO response_cache (model_id, input_hash, parameters, response) VALUES (?, ?, ?, ?)`,
		modelID, hash, params, response,
	)
	return err
}

func (c *SQLiteCache) InvalidateModel(ctx context.Context, modelID string) error {
	_, err := c.db.ExecContext(ctx,
		`DELETE FROM embeddings_cache WHERE model_id = ?`, modelID,
	)
	if err != nil {
		return err
	}
	_, err = c.db.ExecContext(ctx,
		`DELETE FROM response_cache WHERE model_id = ?`, modelID,
	)
	return err
}

func (c *SQLiteCache) Clear(ctx context.Context) error {
	_, err := c.db.ExecContext(ctx, `DELETE FROM embeddings_cache`)
	if err != nil {
		return err
	}
	_, err = c.db.ExecContext(ctx, `DELETE FROM response_cache`)
	return err
}

func (c *SQLiteCache) Close() error {
	return c.db.Close()
}

func (c *SQLiteCache) SaveSession(ctx context.Context, id, modelID, messages, metadata string) error {
	_, err := c.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO sessions (id, model_id, messages, metadata, updated_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		id, modelID, messages, metadata,
	)
	return err
}

func (c *SQLiteCache) GetSession(ctx context.Context, id string) (string, string, string, error) {
	var modelID, messages, metadata string
	err := c.db.QueryRowContext(ctx,
		`SELECT model_id, messages, metadata FROM sessions WHERE id = ?`, id,
	).Scan(&modelID, &messages, &metadata)
	if err != nil {
		return "", "", "", err
	}
	return modelID, messages, metadata, nil
}

func hashInput(input string) string {
	h := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", h[:])
}

func floatsToBytes(f []float32) []byte {
	b := make([]byte, len(f)*4)
	for i, v := range f {
		u := float32ToBytes(v)
		b[i*4] = u[0]
		b[i*4+1] = u[1]
		b[i*4+2] = u[2]
		b[i*4+3] = u[3]
	}
	return b
}

func bytesToFloats(b []byte) []float32 {
	f := make([]float32, len(b)/4)
	for i := range f {
		f[i] = bytesToFloat32(b[i*4 : i*4+4])
	}
	return f
}

func float32ToBytes(f float32) [4]byte {
	v := uint32(0)
	v |= uint32(uint8(uint32(f)>>24)) << 24
	return [4]byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)}
}

func bytesToFloat32(b []byte) float32 {
	return float32(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24)
}

func execAll(db *sql.DB, stmts []string) error {
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
