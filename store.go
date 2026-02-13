package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

// NewStore opens (or creates) the SQLite database at the platform config directory.
func NewStore() (*Store, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("config dir: %w", err)
	}

	dir := filepath.Join(configDir, "sysadmin-gtd")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}

	return NewStoreWithPath(filepath.Join(dir, "tasks.db"))
}

// NewStoreWithPath opens a store at an explicit path (useful for tests with ":memory:").
func NewStoreWithPath(dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			date            TEXT    NOT NULL,
			description     TEXT    NOT NULL,
			priority        TEXT    NOT NULL DEFAULT 'B',
			time_estimate   TEXT    NOT NULL DEFAULT '',
			is_completed    INTEGER NOT NULL DEFAULT 0,
			carried_from_id INTEGER REFERENCES tasks(id)
		)
	`)
	return err
}

func (s *Store) GetTasksForDate(date string) ([]Task, error) {
	rows, err := s.db.Query(
		`SELECT id, date, description, priority, time_estimate, is_completed, carried_from_id
		 FROM tasks WHERE date = ? ORDER BY priority, id`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTasks(rows)
}

func (s *Store) AddTask(date, description string, priority Priority, timeEstimate string) error {
	_, err := s.db.Exec(
		`INSERT INTO tasks (date, description, priority, time_estimate) VALUES (?, ?, ?, ?)`,
		date, description, string(priority), timeEstimate)
	return err
}

func (s *Store) UpdateTask(id int64, description string, priority Priority, timeEstimate string) error {
	_, err := s.db.Exec(
		`UPDATE tasks SET description = ?, priority = ?, time_estimate = ? WHERE id = ?`,
		description, string(priority), timeEstimate, id)
	return err
}

func (s *Store) DeleteTask(id int64) error {
	_, err := s.db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	return err
}

func (s *Store) MarkComplete(id int64) error {
	_, err := s.db.Exec(`UPDATE tasks SET is_completed = 1 WHERE id = ?`, id)
	return err
}

func (s *Store) MarkIncomplete(id int64) error {
	_, err := s.db.Exec(`UPDATE tasks SET is_completed = 0 WHERE id = ?`, id)
	return err
}

func (s *Store) GetTask(id int64) (Task, error) {
	row := s.db.QueryRow(
		`SELECT id, date, description, priority, time_estimate, is_completed, carried_from_id
		 FROM tasks WHERE id = ?`, id)

	var t Task
	var carriedFromID sql.NullInt64
	err := row.Scan(&t.ID, &t.Date, &t.Description, &t.Priority, &t.TimeEstimate, &t.IsCompleted, &carriedFromID)
	if err != nil {
		return Task{}, err
	}
	if carriedFromID.Valid {
		t.CarriedFromID = &carriedFromID.Int64
	}
	return t, nil
}

// GetCarryOverCandidates returns incomplete tasks for fromDate that haven't already
// been carried over to the next day.
func (s *Store) GetCarryOverCandidates(fromDate, toDate string) ([]Task, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.date, t.description, t.priority, t.time_estimate, t.is_completed, t.carried_from_id
		FROM tasks t
		WHERE t.date = ?
		  AND t.is_completed = 0
		  AND t.id NOT IN (
			SELECT carried_from_id FROM tasks WHERE date = ? AND carried_from_id IS NOT NULL
		  )
		ORDER BY t.priority, t.id`, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTasks(rows)
}

// CarryOverTasks creates copies of the given tasks for toDate, setting carried_from_id.
func (s *Store) CarryOverTasks(tasks []Task, toDate string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT INTO tasks (date, description, priority, time_estimate, carried_from_id) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range tasks {
		if _, err := stmt.Exec(toDate, t.Description, string(t.Priority), t.TimeEstimate, t.ID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func scanTasks(rows *sql.Rows) ([]Task, error) {
	var tasks []Task
	for rows.Next() {
		var t Task
		var carriedFromID sql.NullInt64
		if err := rows.Scan(&t.ID, &t.Date, &t.Description, &t.Priority, &t.TimeEstimate, &t.IsCompleted, &carriedFromID); err != nil {
			return nil, err
		}
		if carriedFromID.Valid {
			t.CarriedFromID = &carriedFromID.Int64
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
