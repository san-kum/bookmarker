bookmark-manager/
├── cmd/
│   └── bookmark/
│       └── main.go         # Entry point
├── internal/
│   ├── app/                # Application core
│   │   ├── app.go          # Main application logic
│   │   └── config.go       # Configuration handling
│   ├── model/              # Data models
│   │   ├── bookmark.go     # Bookmark structure
│   │   ├── tag.go          # Tag structure
│   │   └── user.go         # User structure
│   ├── repository/         # Data access layer
│   │   ├── bookmark_repo.go # Bookmark storage operations
│   │   └── db.go           # Database connection handling
│   ├── service/            # Business logic
│   │   ├── bookmark_service.go # Bookmark-related operations
│   │   ├── extractor/      # Content extraction
│   │   │   └── html_extractor.go # HTML content extraction
│   │   └── search/         # Search functionality
│   │       └── search_service.go # Search algorithms
│   └── ui/                 # User interface
│       ├── tui.go          # TUI main component
│       └── views/          # Different UI views
│           ├── bookmark_view.go
│           ├── search_view.go
│           └── tag_view.go
├── pkg/                    # Public packages
│   └── util/               # Utilities
│       ├── logger.go       # Logging utilities
│       └── validator.go    # Input validation
├── go.mod                  # Go modules
└── go.sum                  # Module checksums
