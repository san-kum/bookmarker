package app

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/san-kum/bookmarker/internal/repository"
	"github.com/san-kum/bookmarker/internal/service"
	"github.com/san-kum/bookmarker/internal/service/extractor"
	"github.com/san-kum/bookmarker/internal/service/search"
	"github.com/san-kum/bookmarker/internal/ui"
)

type App struct {
	config        *Config
	database      *repository.Database
	bookmarkRepo  *repository.BookmarkRepository
	bookmarkSvc   *service.BookmarkService
	searchService *search.SearchService
	ui            *ui.TUI
}

func NewApp() (*App, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize configuration: %w", err)
	}

	db, err := repository.NewDatabase(config.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initalize database: %w", err)
	}

	bookmarkRepo := repository.NewBookmarkRepository(db)
	htmlExtractor := extractor.NewHTMLExtractor()
	bookmarkSvc := service.NewBookmarkService(bookmarkRepo, htmlExtractor)

	searchService, err := search.NewSearchService(bookmarkRepo, config.IndexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initalize search service: %w", err)
	}

	tui := ui.NewTUI(bookmarkSvc, searchService)

	return &App{
		config:        config,
		database:      db,
		bookmarkRepo:  bookmarkRepo,
		bookmarkSvc:   bookmarkSvc,
		searchService: searchService,
		ui:            tui,
	}, nil
}

func (a *App) Run() error {
	defer a.cleanup()
	log.Info().Msg("Starting Smart bookmark manager...")
	return a.ui.Run()
}

func (a *App) cleanup() {
	log.Info().Msg("Shutting down application...")
	if err := a.searchService.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close search service")
	}
	if err := a.database.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close database")
	}
}
