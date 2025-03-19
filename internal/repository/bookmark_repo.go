package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/san-kum/bookmarker/internal/model"
)

type BookmarkRepository struct {
	db *Database
}

func NewBookmarkRepository(db *Database) *BookmarkRepository {
	return &BookmarkRepository{
		db: db,
	}
}

func (r *BookmarkRepository) Create(bookmark *model.Bookmark) error {
	tx, err := r.db.GetDB().Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert bookmark
	query := `
    INSERT INTO bookmarks (url, title, description, content, summary, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?)
    `
	res, err := tx.Exec(query, bookmark.URL, bookmark.Title, bookmark.Description, bookmark.Content, bookmark.Summary, bookmark.CreatedAt, bookmark.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert bookmark: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}
	bookmark.ID = id

	for i := range bookmark.Tags {
		tag := &bookmark.Tags[i]
		if tag.ID == 0 {
			tagQuery := `INSERT INTO tags (name) VALUES (?) ON CONFLICT(name) DO UPDATE SET name=name RETURNING id`
			err := tx.Get(&tag.ID, tagQuery, tag.Name)
			if err != nil {
				return fmt.Errorf("failed to insert tag: %w", err)
			}
		}

		_, err = tx.Exec(`INSERT INTO bookmark_tags (bookmark_id, tag_id) VALUES (?, ?)`, id, tag.ID)
		if err != nil {
			return fmt.Errorf("failed to insert bookmark-tag relation: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *BookmarkRepository) GetByID(id int64) (*model.Bookmark, error) {
	var bookmark model.Bookmark

	query := `SELECT * FROM bookmarks WHERE id = ?`
	err := r.db.GetDB().Get(&bookmark, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get bookmark: %w", err)
	}

	// get tags
	tagsQuery := `
  SELECT t.id, t.name
  FROM tags t
  JOIN bookmark_tags bt ON bt.tag_id = t.id
  WHERE bt.bookmark_id = ?
  `

	if err := r.db.GetDB().Select(&bookmark.Tags, tagsQuery, id); err != nil {
		return nil, fmt.Errorf("failed to get bookmark tags: %w", err)
	}

	return &bookmark, nil
}

func (r *BookmarkRepository) GetByURL(url string) (*model.Bookmark, error) {
	var bookmark model.Bookmark

	query := `SELECT * FROM bookmarks WHERE url = ?`
	err := r.db.GetDB().Get(&bookmark, query, url)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get bookmark: %w", err)
	}

	tagsQuery := `
    SELECT t.id, t.name
    FROM tags t
    JOIN bookmark_tags bt ON bt.tag_id = t.id
    WHERE bt.bookmark_id = ?
    `
	if err := r.db.GetDB().Select(&bookmark.Tags, tagsQuery, bookmark.ID); err != nil {
		return nil, fmt.Errorf("failed to get bookmark tags: %w", err)
	}

	return &bookmark, nil
}

func (r *BookmarkRepository) List(tag string, limit, offset int) ([]*model.Bookmark, error) {
	var bookmarks []*model.Bookmark
	var query string
	var args []interface{}

	if tag != "" {
		// filter by tag
		query = `
     SELECT b.*
      FROM bookmarks b
      JOIN bookmark_tags bt ON bt.bookmark_id = b.id
      JOIN tags t ON t.id = bt.tag_id
      WHERE t.name = ?
      ORDER BY b.created_at DESC
      LIMIT ? OFFSET ?
      `
		args = []interface{}{tag, limit, offset}
	} else {
		// get all bookmarks
		query = `
    SELECT * FROM bookmarks
    ORDER BY created_at DESC
    LIMIT ? OFFSET ?
    `
		args = []interface{}{limit, offset}
	}

	if err := r.db.GetDB().Select(&bookmarks, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list bookmarks: %w", err)
	}

	for _, bookmark := range bookmarks {
		tagsQuery := `
    SELECT t.id, t.name
    FROM tags t
    JOIN bookmark_tags bt ON bt.tag_id = t.id
    WHERE bt.bookmark_id = ?
    `
		if err := r.db.GetDB().Select(&bookmark.Tags, tagsQuery, bookmark.ID); err != nil {
			return nil, fmt.Errorf("failed to get bookmark tags: %w", err)
		}
	}

	return bookmarks, nil

}

func (r *BookmarkRepository) Update(bookmark *model.Bookmark) error {
	tx, err := r.db.GetDB().Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	bookmark.UpdatedAt = time.Now()
	query := `
   UPDATE bookmarks
   SET url = ?, title = ?, description = ?, content = ?, summary = ?, updated_at = ?
   WHERE id = ?
  `
	_, err = tx.Exec(query, bookmark.URL, bookmark.Title, bookmark.Description, bookmark.Content, bookmark.Summary, bookmark.UpdatedAt, bookmark.ID)
	if err != nil {
		return fmt.Errorf("failed to update bookmark: %w", err)
	}

	_, err = tx.Exec(`DELETE FROM bookmark_tags WHERE bookmark_id = ?`, bookmark.ID)
	if err != nil {
		return fmt.Errorf("failed to delete bookmark-tag relations: %w", err)
	}

	for i := range bookmark.Tags {
		tag := &bookmark.Tags[i]
		if tag.ID == 0 {
			tagQuery := `INSERT INTO tags (name) VALUES (?) ON CONFLICT(name) DO UPDATE SET name=name repository id`
			err := tx.Get(&tag.ID, tagQuery, tag.Name)
			if err != nil {
				return fmt.Errorf("failed to insert tag: %w", err)
			}
		}

		_, err = tx.Exec(`INSERT INTO bookmark_tags (bookmark_id, tag_id) VALUES (?, ?)`, bookmark.ID, tag.ID)
		if err != nil {
			return fmt.Errorf("failed to insert bookmark-tag relation: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *BookmarkRepository) Delete(id int64) error {
	_, err := r.db.GetDB().Exec(`DELETE FROM bookmarks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}
	return nil
}

func (r *BookmarkRepository) GetAllTags() ([]model.Tag, error) {
	var tags []model.Tag
	query := `SELECT * FROM tags ORDER BY name`
	if err := r.db.GetDB().Select(&tags, query); err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	return tags, nil
}
