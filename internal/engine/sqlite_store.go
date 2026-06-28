package engine

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteSessionStore stores sessions in a SQLite database.
type SQLiteSessionStore struct {
	db *sql.DB
}

func NewSQLiteSessionStore(path string) (*SQLiteSessionStore, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("sqlite ping: %w", err)
	}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id         TEXT PRIMARY KEY,
			model_id   TEXT NOT NULL,
			messages   TEXT NOT NULL DEFAULT '[]',
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			max_tokens INTEGER DEFAULT 0
		)
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite init: %w", err)
	}
	return &SQLiteSessionStore{db: db}, nil
}

func (s *SQLiteSessionStore) Create(_ context.Context, session *Session) error {
	if session.ID == "" {
		return fmt.Errorf("session ID is required")
	}
	msgData, err := json.Marshal(session.Messages)
	if err != nil {
		return fmt.Errorf("marshal messages: %w", err)
	}
	_, err = s.db.Exec(
		`INSERT INTO sessions (id, model_id, messages, created_at, updated_at, max_tokens) VALUES (?, ?, ?, ?, ?, ?)`,
		session.ID, session.ModelID, string(msgData),
		session.CreatedAt.UTC().Format(time.RFC3339Nano),
		session.UpdatedAt.UTC().Format(time.RFC3339Nano),
		session.MaxTokens,
	)
	if err != nil {
		return fmt.Errorf("sqlite insert: %w", err)
	}
	return nil
}

func (s *SQLiteSessionStore) Get(_ context.Context, id string) (*Session, error) {
	var session Session
	var msgJSON string
	var createdStr, updatedStr string
	err := s.db.QueryRow(
		`SELECT id, model_id, messages, created_at, updated_at, max_tokens FROM sessions WHERE id = ?`, id,
	).Scan(&session.ID, &session.ModelID, &msgJSON, &createdStr, &updatedStr, &session.MaxTokens)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("sqlite query: %w", err)
	}
	if err := json.Unmarshal([]byte(msgJSON), &session.Messages); err != nil {
		return nil, fmt.Errorf("unmarshal messages: %w", err)
	}
	session.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
	session.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
	return &session, nil
}

func (s *SQLiteSessionStore) Update(_ context.Context, session *Session) error {
	msgData, err := json.Marshal(session.Messages)
	if err != nil {
		return fmt.Errorf("marshal messages: %w", err)
	}
	res, err := s.db.Exec(
		`UPDATE sessions SET model_id = ?, messages = ?, updated_at = ?, max_tokens = ? WHERE id = ?`,
		session.ModelID, string(msgData),
		session.UpdatedAt.UTC().Format(time.RFC3339Nano),
		session.MaxTokens, session.ID,
	)
	if err != nil {
		return fmt.Errorf("sqlite update: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("session not found: %s", session.ID)
	}
	return nil
}

func (s *SQLiteSessionStore) Delete(_ context.Context, id string) error {
	res, err := s.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("sqlite delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("session not found: %s", id)
	}
	return nil
}

func (s *SQLiteSessionStore) List(_ context.Context) ([]*Session, error) {
	rows, err := s.db.Query(
		`SELECT id, model_id, messages, created_at, updated_at, max_tokens FROM sessions ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("sqlite list: %w", err)
	}
	defer rows.Close()

	var result []*Session
	for rows.Next() {
		var session Session
		var msgJSON string
		var createdStr, updatedStr string
		if err := rows.Scan(&session.ID, &session.ModelID, &msgJSON, &createdStr, &updatedStr, &session.MaxTokens); err != nil {
			return nil, fmt.Errorf("sqlite scan: %w", err)
		}
		json.Unmarshal([]byte(msgJSON), &session.Messages)
		session.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
		session.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
		result = append(result, &session)
	}
	return result, rows.Err()
}

func (s *SQLiteSessionStore) Close(_ context.Context) error {
	return s.db.Close()
}
