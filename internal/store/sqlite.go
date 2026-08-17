package store

import (
	"ctx/internal/model"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore() (*SQLiteStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".ctx")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dir, "ctx.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return &SQLiteStore{db: db}, nil
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS context_entries (
		id          TEXT PRIMARY KEY,
		repo        TEXT NOT NULL,
		branch      TEXT NOT NULL,
		commit_sha  TEXT,
		files       TEXT,      -- JSON array as text
		session     TEXT,
		completed   TEXT,
		remaining   TEXT,
		blocker     TEXT,
		type        TEXT,
		created_at  DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_repo ON context_entries(repo);
	`
	_, err := db.Exec(schema)
	return err
}

func (s *SQLiteStore) Save(entry model.ContextEntry) error {
	filesJSON, err := json.Marshal(entry.Files)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		INSERT INTO context_entries
			(id, repo, branch, commit_sha, files, session, completed, remaining, blocker, type, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.Repo, entry.Branch, entry.CommitSHA, string(filesJSON),
		entry.Session, entry.Completed, entry.Remaining, entry.Blocker,
		string(entry.Type), entry.CreatedAt,
	)
	return err
}

func (s *SQLiteStore) LastByRepo(repo string) (*model.ContextEntry, error) {
	row := s.db.QueryRow(`
		SELECT id, repo, branch, commit_sha, files, session, completed, remaining, blocker, type, created_at
		FROM context_entries
		WHERE repo = ?
		ORDER BY created_at DESC
		LIMIT 1`, repo)

	entry, err := scanEntry(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *SQLiteStore) AllByRepo(repo string) ([]model.ContextEntry, error) {
	rows, err := s.db.Query(`
		SELECT id, repo, branch, commit_sha, files, session, completed, remaining, blocker, type, created_at
		FROM context_entries
		WHERE repo = ?
		ORDER BY created_at DESC`, repo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.ContextEntry
	for rows.Next() {
		var e model.ContextEntry
		var filesJSON string
		var typeStr string

		if err := rows.Scan(&e.ID, &e.Repo, &e.Branch, &e.CommitSHA, &filesJSON,
			&e.Session, &e.Completed, &e.Remaining, &e.Blocker, &typeStr, &e.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(filesJSON), &e.Files)
		e.Type = model.EntryType(typeStr)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func scanEntry(row *sql.Row) (*model.ContextEntry, error) {
	var e model.ContextEntry
	var filesJSON string
	var typeStr string
	var createdAt time.Time

	err := row.Scan(&e.ID, &e.Repo, &e.Branch, &e.CommitSHA, &filesJSON,
		&e.Session, &e.Completed, &e.Remaining, &e.Blocker, &typeStr, &createdAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(filesJSON), &e.Files)
	e.Type = model.EntryType(typeStr)
	e.CreatedAt = createdAt
	return &e, nil
}
