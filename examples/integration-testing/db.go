package main

import (
	"context"
	"database/sql"
	"encoding/json"

	_ "modernc.org/sqlite"
)

type Todo struct {
	ID    int64
	Title string
	Done  bool
}

type DB struct {
	db *sql.DB
}

func NewDB() (*DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`
		CREATE TABLE todos (
			id    INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT    NOT NULL,
			done  INTEGER NOT NULL DEFAULT 0
		);
	`); err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

func (s *DB) List(ctx context.Context) ([]Todo, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, done FROM todos ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Done); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *DB) Create(ctx context.Context, title string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO todos (title) VALUES (?)`, title)
	return err
}

func (s *DB) Toggle(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE todos SET done = 1 - done WHERE id = ?`, id)
	return err
}

func (s *DB) Delete(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM todos WHERE id = ?`, id)
	return err
}

func (s *DB) DumpRows(ctx context.Context) ([]string, error) {
	tables, err := s.db.QueryContext(ctx,
		`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return nil, err
	}
	var names []string
	for tables.Next() {
		var n string
		if err := tables.Scan(&n); err != nil {
			tables.Close()
			return nil, err
		}
		names = append(names, n)
	}
	tables.Close()

	var out []string
	for _, table := range names {
		rows, err := s.db.QueryContext(ctx, `SELECT * FROM `+table+` ORDER BY ROWID`)
		if err != nil {
			return nil, err
		}
		cols, err := rows.Columns()
		if err != nil {
			rows.Close()
			return nil, err
		}
		for rows.Next() {
			vals := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				rows.Close()
				return nil, err
			}
			row := map[string]any{"_table": table}
			for i, c := range cols {
				v := vals[i]
				if b, ok := v.([]byte); ok {
					v = string(b)
				}
				row[c] = v
			}
			b, err := json.Marshal(row)
			if err != nil {
				rows.Close()
				return nil, err
			}
			out = append(out, string(b))
		}
		rows.Close()
	}
	return out, nil
}
