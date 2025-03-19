package model

import "time"

type Bookmark struct {
	ID          int64     `db:"id" json:"id"`
	URL         string    `db:"url" json:"url"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	Content     string    `db:"content" json:"content"`
	Summary     string    `db:"summary" json:"summary"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
	Tags        []Tag     `json:"tags"`
}

func NewBookmark(url, title string) *Bookmark {
	now := time.Now()
	return &Bookmark{
		URL:       url,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
		Tags:      make([]Tag, 0),
	}
}

func (b *Bookmark) AddTag(tag Tag) {
	for _, t := range b.Tags {
		if t.Name == tag.Name {
			return
		}
	}
	b.Tags = append(b.Tags, tag)
}

func (b *Bookmark) RemoveTag(tagName string) {
	for i, tag := range b.Tags {
		if tag.Name == tagName {
			b.Tags = append(b.Tags[:i], b.Tags[i+1:]...)
			return
		}
	}
}
