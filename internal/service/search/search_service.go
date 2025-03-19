package search

import (
	"fmt"

	"github.com/blevesearch/bleve"
	"github.com/rs/zerolog/log"
	"github.com/san-kum/bookmarker/internal/model"
	"github.com/san-kum/bookmarker/internal/repository"
)

type BookmarkIndex struct {
	ID          string   `json:"id"`
	URL         string   `json:"url"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Summary     string   `json:"summary"`
	Tags        []string `json:"tags"`
}

type SearchService struct {
	repo      *repository.BookmarkRepository
	index     bleve.Index
	indexPath string
}

func NewSearchService(repo *repository.BookmarkRepository, indexPath string) (*SearchService, error) {
	index, err := openOrCreateIndex(indexPath)

	if err != nil {
		return nil, err
	}

	service := &SearchService{
		repo:      repo,
		index:     index,
		indexPath: indexPath,
	}

	return service, nil
}

func openOrCreateIndex(indexPath string) (bleve.Index, error) {
	index, err := bleve.Open(indexPath)
	if err == bleve.ErrorIndexPathDoesNotExist {
		mapping := bleve.NewIndexMapping()

		index, err = bleve.New(indexPath, mapping)
		if err != nil {
			return nil, fmt.Errorf("failed to create search index: %w", err)
		}
		log.Info().Msg("Created new search index")
	} else if err != nil {
		return nil, fmt.Errorf("failed to open search index: %w", err)
	} else {
		log.Info().Msg("Opened existing search index")
	}
	return index, nil
}

func (s *SearchService) Close() error {
	return s.index.Close()
}

func (s *SearchService) IndexBookmark(bookmark *model.Bookmark) error {
	tagNames := make([]string, len(bookmark.Tags))
	for i, tag := range bookmark.Tags {
		tagNames[i] = tag.Name
	}
	doc := BookmarkIndex{
		ID:          fmt.Sprintf("%d", bookmark.ID),
		URL:         bookmark.URL,
		Title:       bookmark.Title,
		Description: bookmark.Description,
		Content:     bookmark.Content,
		Summary:     bookmark.Summary,
		Tags:        tagNames,
	}
	return s.index.Index(doc.ID, doc)
}

func (s *SearchService) DeleteBookmark(id int64) error {
	return s.index.Delete(fmt.Sprintf("%d", id))
}

func (s *SearchService) Search(query string, limit int) ([]*model.Bookmark, error) {
	if limit <= 0 {
		limit = 20
	}

	searchQuery := bleve.NewQueryStringQuery(query)
	searchRequest := bleve.NewSearchRequest(searchQuery)
	searchRequest.Size = limit
	searchRequest.Fields = []string{"id"}

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	var bookmarks []*model.Bookmark
	for _, hit := range searchResults.Hits {
		idStr, ok := hit.Fields["id"].(string)
		if !ok {
			log.Warn().Str("docID", hit.ID).Msg("Invalid ID field in search result")
			continue
		}

		var id int64
		_, err := fmt.Scanf(idStr, "%d", &id)
		if err != nil {
			log.Warn().Str("docID", hit.ID).Err(err).Msg("Failed to parse bookmark ID")
			continue
		}

		bookmark, err := s.repo.GetByID(id)
		if err != nil {
			log.Warn().Int64("id", id).Err(err).Msg("Failed to featch bookmark data")
			continue
		}

		if bookmark != nil {
			bookmarks = append(bookmarks, bookmark)
		}
	}

	return bookmarks, nil

}

func (s *SearchService) RebuildIndex() error {
	bookmarks, err := s.repo.List("", 1000, 0)
	if err != nil {
		return fmt.Errorf("failed to fetch bookmarks: %w", err)
	}

	err = s.index.Close()
	if err != nil {
		return fmt.Errorf("failed to close index: %w", err)
	}

	s.index, err = openOrCreateIndex(s.indexPath)
	if err != nil {
		return fmt.Errorf("failed to recreate index: %w", err)
	}

	batch := s.index.NewBatch()
	for _, bookmark := range bookmarks {
		tagNames := make([]string, len(bookmark.Tags))
		for i, tag := range bookmark.Tags {
			tagNames[i] = tag.Name
		}

		doc := BookmarkIndex{
			ID:          fmt.Sprintf("%d", bookmark.ID),
			URL:         bookmark.URL,
			Title:       bookmark.Title,
			Description: bookmark.Description,
			Content:     bookmark.Content,
			Summary:     bookmark.Summary,
			Tags:        tagNames,
		}

		err = batch.Index(doc.ID, doc)
		if err != nil {
			return fmt.Errorf("failed to add document to batch: %w", err)
		}
	}

	err = s.index.Batch(batch)
	if err != nil {
		return fmt.Errorf("failed to execute batch: %w", err)
	}

	log.Info().Int("count", len(bookmarks)).Msg("Search index rebuilt successfully")
	return nil
}
