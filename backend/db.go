package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const defaultResumeDBPath = "data/resumes.db"
const defaultSeedResumePath = "resume.json"
const defaultSeedResumeName = "FAANG Resume Sample"

var resumeStore *ResumeStore

type ResumeStore struct {
	db *sql.DB
}

type ResumeSummary struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ResumeRecord struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Resume    Resume    `json:"resume"`
}

func initResumeStore(dbPath, seedPath string) error {
	if strings.TrimSpace(dbPath) == "" {
		dbPath = defaultResumeDBPath
	}
	if strings.TrimSpace(seedPath) == "" {
		seedPath = defaultSeedResumePath
	}

	_, statErr := os.Stat(dbPath)
	wasMissing := os.IsNotExist(statErr)

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open sqlite db: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(0)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return fmt.Errorf("ping sqlite db: %w", err)
	}

	store := &ResumeStore{db: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return err
	}

	if err := store.seedDefaultResume(wasMissing, seedPath); err != nil {
		_ = db.Close()
		return err
	}

	resumeStore = store
	return nil
}

func (s *ResumeStore) seedDefaultResume(wasMissing bool, seedPath string) error {
	count, err := s.countResumes()
	if err != nil {
		return err
	}

	// Seed only when the database is brand new or contains no resumes.
	if !wasMissing && count > 0 {
		return nil
	}

	if count > 0 {
		return nil
	}

	data, err := os.ReadFile(seedPath)
	if err != nil {
		return fmt.Errorf("read seed resume %q: %w", seedPath, err)
	}

	resume, err := parseResume(data)
	if err != nil {
		return fmt.Errorf("parse seed resume %q: %w", seedPath, err)
	}

	if _, err := s.Create(defaultSeedResumeName, resume); err != nil {
		return fmt.Errorf("seed default resume: %w", err)
	}

	return nil
}

func (s *ResumeStore) countResumes() (int64, error) {
	row := s.db.QueryRow(`SELECT COUNT(*) FROM resumes`)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count resumes: %w", err)
	}
	return count, nil
}

func (s *ResumeStore) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS resumes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	payload TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_resumes_created_at ON resumes(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_resumes_updated_at ON resumes(updated_at DESC);
`
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("migrate sqlite schema: %w", err)
	}
	return nil
}

func (s *ResumeStore) List() ([]ResumeSummary, error) {
	rows, err := s.db.Query(`
		SELECT id, name, created_at, updated_at
		FROM resumes
		ORDER BY datetime(created_at) DESC, id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list resumes: %w", err)
	}
	defer rows.Close()

	var resumes []ResumeSummary
	for rows.Next() {
		var summary ResumeSummary
		var createdAt string
		var updatedAt string
		if err := rows.Scan(&summary.ID, &summary.Name, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan resume summary: %w", err)
		}

		summary.CreatedAt, err = parseTimestamp(createdAt)
		if err != nil {
			return nil, err
		}
		summary.UpdatedAt, err = parseTimestamp(updatedAt)
		if err != nil {
			return nil, err
		}
		resumes = append(resumes, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate resume summaries: %w", err)
	}

	return resumes, nil
}

func (s *ResumeStore) Get(id int64) (ResumeRecord, error) {
	row := s.db.QueryRow(`
		SELECT id, name, created_at, updated_at, payload
		FROM resumes
		WHERE id = ?
	`, id)

	var record ResumeRecord
	var payload string
	var createdAt string
	var updatedAt string
	var err error
	if err := row.Scan(&record.ID, &record.Name, &createdAt, &updatedAt, &payload); err != nil {
		return ResumeRecord{}, mapResumeNotFound(err, id)
	}

	if err := json.Unmarshal([]byte(payload), &record.Resume); err != nil {
		return ResumeRecord{}, fmt.Errorf("unmarshal resume payload: %w", err)
	}

	record.CreatedAt, err = parseTimestamp(createdAt)
	if err != nil {
		return ResumeRecord{}, err
	}
	record.UpdatedAt, err = parseTimestamp(updatedAt)
	if err != nil {
		return ResumeRecord{}, err
	}

	return record, nil
}

func (s *ResumeStore) Create(name string, resume Resume) (ResumeRecord, error) {
	if strings.TrimSpace(name) == "" {
		name = "Untitled Resume"
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	payload, err := json.Marshal(resume)
	if err != nil {
		return ResumeRecord{}, fmt.Errorf("marshal resume payload: %w", err)
	}

	result, err := s.db.Exec(`
		INSERT INTO resumes (name, created_at, updated_at, payload)
		VALUES (?, ?, ?, ?)
	`, name, now, now, string(payload))
	if err != nil {
		return ResumeRecord{}, fmt.Errorf("insert resume: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return ResumeRecord{}, fmt.Errorf("last insert id: %w", err)
	}

	return s.Get(id)
}

func (s *ResumeStore) Update(id int64, name string, resume Resume) (ResumeRecord, error) {
	if strings.TrimSpace(name) == "" {
		name = "Untitled Resume"
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	payload, err := json.Marshal(resume)
	if err != nil {
		return ResumeRecord{}, fmt.Errorf("marshal resume payload: %w", err)
	}

	result, err := s.db.Exec(`
		UPDATE resumes
		SET name = ?, updated_at = ?, payload = ?
		WHERE id = ?
	`, name, now, string(payload), id)
	if err != nil {
		return ResumeRecord{}, fmt.Errorf("update resume: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return ResumeRecord{}, fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ResumeRecord{}, fmt.Errorf("resume %d not found", id)
	}

	return s.Get(id)
}

func (s *ResumeStore) Delete(id int64) error {
	result, err := s.db.Exec(`DELETE FROM resumes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete resume: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("resume %d not found", id)
	}

	return nil
}

func parseTimestamp(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err == nil {
		return parsed, nil
	}

	parsed, err = time.Parse(time.RFC3339, value)
	if err == nil {
		return parsed, nil
	}

	return time.Time{}, fmt.Errorf("parse timestamp %q: %w", value, err)
}

func mapResumeNotFound(err error, id int64) error {
	if err == sql.ErrNoRows {
		return fmt.Errorf("resume %d not found", id)
	}
	return fmt.Errorf("get resume %d: %w", id, err)
}
