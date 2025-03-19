package service

import (
	"fmt"
	"net/url"

	"github.com/rs/zerolog/log"
	"github.com/san-kum/bookmarker/internal/model"
	"github.com/san-kum/bookmarker/internal/repository"
	"github.com/san-kum/bookmarker/internal/service/extractor"
)

type BookmarkService struct {
	repo      *repository.BookmarkRepository
	extractor *extractor.HTMLExtractor
}

func NewBookmarkService(repo *repository.BookmarkRepository, extractor *extractor.HTMLExtractor) *BookmarkService {
	return &BookmarkService{
		repo:      repo,
		extractor: extractor,
	}
}

func (s *BookmarkService) Add(urlStr string, tags []string) (*model.Bookmark, error) {
	_, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	existing, err := s.repo.GetByURL(urlStr)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	title, description, content, err := s.extractor.ExtractContent(urlStr)
	if err != nil {
		log.Warn().Err(err).Str("url", urlStr).Msg("Content extraction failed, creating bookmark with minimal info")
		bookmark := model.NewBookmark(urlStr, urlStr)
		err = s.repo.Create(bookmark)
		if err != nil {
			return nil, err
		}
		return bookmark, nil
	}

	summary := s.extractor.GenerateSummary(content)

	bookmark := model.NewBookmark(urlStr, title)
	bookmark.Description = description
	bookmark.Content = content
	bookmark.Summary = summary

	for _, tagName := range tags {
		if tagName != "" {
			bookmark.AddTag(model.NewTag(tagName))
		}
	}

	err = s.repo.Create(bookmark)
	if err != nil {
		return nil, err
	}

	return bookmark, nil
}

func (s *BookmarkService) Get(id int64) (*model.Bookmark, error) {
	return s.repo.GetByID(id)
}

func (s *BookmarkService) List(tag string, limit, offset int) ([]*model.Bookmark, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.List(tag, limit, offset)
}

func (s *BookmarkService) Update(bookmark *model.Bookmark) error {
	return s.repo.Update(bookmark)
}

func (s *BookmarkService) Delete(id int64) error {
	return s.repo.Delete(id)
}

func (s *BookmarkService) AddTag(bookmarkID int64, tagName string) error {
	bookmark, err := s.repo.GetByID(bookmarkID)
	if err != nil {
		return err
	}
	if bookmark == nil {
		return fmt.Errorf("bookmark not found.")
	}
	bookmark.AddTag(model.NewTag(tagName))
	return s.repo.Update(bookmark)
}

func (s *BookmarkService) RemoveTag(bookmarkID int64, tagName string) error {
	bookmark, err := s.repo.GetByID(bookmarkID)
	if err != nil {
		return err
	}

	if bookmark == nil {
		return fmt.Errorf("bookmark not found.")
	}

	bookmark.RemoveTag(tagName)
	return s.repo.Update(bookmark)
}

func (s *BookmarkService) GetAllTags() ([]model.Tag, error) {
	return s.repo.GetAllTags()
}
