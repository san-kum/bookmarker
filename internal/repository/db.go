package repository

import (
	"fmt"
	"os"
	"path/filepath"
  _ "github.com/mattn/go-sqlite3"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type Database struct {
	db *sqlx.DB
}

func NewDatabase(dataDir string) (*Database, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "bookmarks.db")
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := initScheme(db); err != nil {
		return nil, fmt.Errorf("failed to initialize database scheme: %w", err)
	}
	return &Database{db: db}, nil

}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) GetDB() *sqlx.DB {
	return d.db
}

func initScheme(db *sqlx.DB) error {
	_, err := db.Exec(`
  CREATE TABLE IF NOT EXISTS bookmarks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        url TEXT UNIQUE NOT NULL,
        title TEXT,
        description TEXT,
        content TEXT,
        summary TEXT,
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP NOT NULL
      );
  `)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
  CREATE TABLE IF NOT EXISTS tags (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT UNIQUE NOT NULL
      );
  `)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
  CREATE TABLE IF NOT EXISTS bookmark_tags (
        bookmark_id INTEGER,
        tag_id INTEGER,
        PRIMARY KEY (bookmark_id, tag_id),
        FOREIGN KEY (bookmark_id) REFERENCES bookmarks(id) ON DELETE CASCADE,
        FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
      );
  `)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
   CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL,
        created_at TIMESTAMP NOT NULL
      );
  `)
	if err != nil {
		return err
	}
	log.Info().Msg("Database schema initialized successfully.")
	return nil
}




