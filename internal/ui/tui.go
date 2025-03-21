package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/san-kum/bookmarker/internal/model"
	"github.com/san-kum/bookmarker/internal/service"
	"github.com/san-kum/bookmarker/internal/service/search"
)

// TUI represents the text user interface
type TUI struct {
	app             *tview.Application
	pages           *tview.Pages
	bookmarkService *service.BookmarkService
	searchService   *search.SearchService

	// Pages
	mainPage         *tview.Flex
	bookmarkListPage *tview.Flex
	searchPage       *tview.Flex
	addBookmarkPage  *tview.Flex
	viewBookmarkPage *tview.Flex

	// Components
	bookmarkList *tview.List
	statusBar    *tview.TextView
	helpBar      *tview.TextView

	// State
	currentBookmarks []*model.Bookmark
	currentTags      []model.Tag

	filterInput     *tview.InputField
	addBookmarkForm *tview.Form
	urlInput        *tview.InputField
	tagsInput       *tview.InputField
}

// NewTUI creates a new TUI
func NewTUI(bookmarkService *service.BookmarkService, searchService *search.SearchService) *TUI {
	tui := &TUI{
		app:             tview.NewApplication(),
		bookmarkService: bookmarkService,
		searchService:   searchService,
	}

	// Initialize UI components
	tui.setupUI()

	return tui
}

// setupUI initializes all UI components
func (t *TUI) setupUI() {
	// Create pages
	t.pages = tview.NewPages()

	// Set up components
	t.setupStatusBar()
	t.setupHelpBar()
	t.setupMainPage()
	t.setupBookmarkListPage()
	t.setupSearchPage()
	t.setupAddBookmarkPage()
	t.setupViewBookmarkPage()

	// Add pages
	t.pages.AddPage("main", t.mainPage, true, true)
	t.pages.AddPage("bookmarkList", t.bookmarkListPage, true, false)
	t.pages.AddPage("search", t.searchPage, true, false)
	t.pages.AddPage("addBookmark", t.addBookmarkPage, true, false)
	t.pages.AddPage("viewBookmark", t.viewBookmarkPage, true, false)

	// Set global keyboard handlers
	t.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			// Go back to main page from any page
			currentPage, _ := t.pages.GetFrontPage()
			if currentPage != "main" {
				t.showPage("main")
				return nil
			}
		case tcell.KeyCtrlQ:
			// Quit application
			t.app.Stop()
			return nil
		}
		return event
	})
}

// setupStatusBar initializes the status bar
func (t *TUI) setupStatusBar() {
	t.statusBar = tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)
	t.statusBar.SetBorder(true).SetTitle(" Status ")
	t.statusBar.SetText("[green]Ready[white]")
}

// setupHelpBar initializes the help bar
func (t *TUI) setupHelpBar() {
	t.helpBar = tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)
	t.helpBar.SetBorder(true).SetTitle(" Help ")
	t.helpBar.SetText("[yellow]Ctrl+Q[white]: Quit | [yellow]Esc[white]: Back | [yellow]Enter[white]: Select")
}

// setupMainPage initializes the main menu page
func (t *TUI) setupMainPage() {
	// Create menu
	menu := tview.NewList().
		AddItem("List Bookmarks", "View and manage your bookmarks", 'l', func() {
			t.loadBookmarks("")
			t.showPage("bookmarkList")
		}).
		AddItem("Search", "Search your bookmarks", 's', func() {
			t.showPage("search")
		}).
		AddItem("Add Bookmark", "Add a new bookmark", 'a', func() {
			t.showPage("addBookmark")
		}).
		AddItem("Quit", "Exit the application", 'q', func() {
			t.app.Stop()
		})

	menu.SetBorder(true).SetTitle(" Smart Bookmark Manager ")

	// Create layout
	t.mainPage = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(menu, 0, 1, true).
		AddItem(t.statusBar, 1, 0, false).
		AddItem(t.helpBar, 1, 0, false)
}

// setupBookmarkListPage initializes the bookmark list page
func (t *TUI) setupBookmarkListPage() {
	// Create bookmark list
	t.bookmarkList = tview.NewList().
		SetSecondaryTextColor(tcell.ColorDimGray)
	t.bookmarkList.SetBorder(true).SetTitle(" Bookmarks ")

	// Create filter input field
	filterInput := tview.NewInputField().
		SetLabel("Filter by tag: ").
		SetFieldWidth(20).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				tag := t.filterInput.GetText()
				t.loadBookmarks(tag)
			}
		})

	// Create tag list
	tagList := tview.NewList().
		SetSecondaryTextColor(tcell.ColorDimGray)
	tagList.SetBorder(true).SetTitle(" Tags ")

	// Load tags
	t.loadTags(tagList)

	// Create right panel with filter and tags
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(filterInput, 1, 0, false).
		AddItem(tagList, 0, 1, false)

	// Create layout
	t.bookmarkListPage = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(t.bookmarkList, 0, 3, true).
			AddItem(rightPanel, 0, 1, false),
			0, 1, true).
		AddItem(t.statusBar, 1, 0, false).
		AddItem(t.helpBar, 1, 0, false)

	// Set bookmark list selected function
	t.bookmarkList.SetSelectedFunc(func(index int, _ string, _ string, _ rune) {
		if index >= 0 && index < len(t.currentBookmarks) {
			t.viewBookmark(t.currentBookmarks[index])
		}
	})

	// Set tag list selected function
	tagList.SetSelectedFunc(func(index int, mainText string, _ string, _ rune) {
		t.loadBookmarks(mainText)
		t.filterInput.SetText(mainText)
	})
}

func (t *TUI) setupSearchPage() {
	// Create search input
	searchInput := tview.NewInputField().
		SetLabel("Search: ").
		SetFieldWidth(40)

	// Create search results list
	searchResults := tview.NewList()
	searchResults.SetBorder(true).SetTitle(" Search Results ")

	searchInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			query := searchInput.GetText()
			t.search(query, searchResults)
			// t.app.SetFocus(searchResults)
		}
	})

	t.searchPage = tview.NewFlex().SetDirection(tview.FlexRow).AddItem(searchInput, 3, 0, true).
		AddItem(searchResults, 0, 2, false).
		AddItem(t.statusBar, 1, 0, false).
		AddItem(t.helpBar, 1, 0, false)

	// Set search results selected function
	searchResults.SetSelectedFunc(func(index int, _ string, _ string, _ rune) {
		if index >= 0 && index < len(t.currentBookmarks) {
			t.viewBookmark(t.currentBookmarks[index])
		}
	})
}

func (t *TUI) setupAddBookmarkPage() {
	t.urlInput = tview.NewInputField().SetLabel("URL").SetFieldWidth(40)
	t.tagsInput = tview.NewInputField().SetLabel("Tags (comma separated)").SetFieldWidth(40)
	t.addBookmarkForm = tview.NewForm().
		AddFormItem(t.urlInput).
		AddFormItem(t.tagsInput).
		AddButton("Add", func() {
			// Get values directly from the stored field references
			url := t.urlInput.GetText()
			tagsStr := t.tagsInput.GetText()

			// Process tags
			tags := []string{}
			if tagsStr != "" {
				tags = strings.Split(tagsStr, ",")
				for i, tag := range tags {
					tags[i] = strings.TrimSpace(tag)
				}
			}

			// Add bookmark
			t.addBookmark(url, tags)

			// Clear form
			t.urlInput.SetText("")
			t.tagsInput.SetText("")
		}).
		AddButton("Cancel", func() {
			t.showPage("main")
		})

	t.addBookmarkForm.SetBorder(true).SetTitle(" Add Bookmark ")

	t.addBookmarkPage = tview.NewFlex().SetDirection(tview.FlexRow).AddItem(nil, 0, 1, false).AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).AddItem(nil, 0, 1, false).AddItem(t.addBookmarkForm, 0, 2, true).AddItem(nil, 0, 1, false), 0, 2, true).AddItem(nil, 0, 1, false).AddItem(t.statusBar, 1, 0, false).AddItem(t.helpBar, 1, 0, false)

}

// setupViewBookmarkPage initializes the view bookmark page
func (t *TUI) setupViewBookmarkPage() {
	// Create text view for bookmark details
	bookmarkDetails := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)
	bookmarkDetails.SetBorder(true).SetTitle(" Bookmark Details ")

	// Create content view
	contentView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)
	contentView.SetBorder(true).SetTitle(" Content ")

	// Create button bar
	buttonBar := tview.NewFlex().SetDirection(tview.FlexColumn)

	// Add buttons
	openButton := tview.NewButton("Open").SetSelectedFunc(func() {
		// This would launch the URL in a browser in a real app
		t.setStatus("[yellow]Opening URL is not implemented in this demo[white]")
	})
	deleteButton := tview.NewButton("Delete").SetSelectedFunc(func() {
		// This would show a confirmation dialog and delete the bookmark
		t.setStatus("[red]Delete not implemented in this demo[white]")
	})
	editTagsButton := tview.NewButton("Edit Tags").SetSelectedFunc(func() {
		// This would show a dialog to edit tags
		t.setStatus("[yellow]Edit tags not implemented in this demo[white]")
	})
	backButton := tview.NewButton("Back").SetSelectedFunc(func() {
		t.showPage("bookmarkList")
	})

	// Add buttons to button bar
	buttonBar.AddItem(openButton, 0, 1, true).
		AddItem(deleteButton, 0, 1, false).
		AddItem(editTagsButton, 0, 1, false).
		AddItem(backButton, 0, 1, false)

	// Create layout
	t.viewBookmarkPage = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(bookmarkDetails, 6, 0, true).
		AddItem(contentView, 0, 1, false).
		AddItem(buttonBar, 1, 0, false).
		AddItem(t.statusBar, 1, 0, false).
		AddItem(t.helpBar, 1, 0, false)
}

// loadBookmarks loads and displays bookmarks
func (t *TUI) loadBookmarks(tag string) {
	var err error

	// Clear list
	t.bookmarkList.Clear()

	// Load bookmarks
	t.currentBookmarks, err = t.bookmarkService.List(tag, 100, 0)
	if err != nil {
		t.setStatus(fmt.Sprintf("[red]Failed to load bookmarks: %v[white]", err))
		return
	}

	// Add bookmarks to list
	for _, bookmark := range t.currentBookmarks {
		title := bookmark.Title
		if title == "" {
			title = bookmark.URL
		}

		// Create secondary text with tags
		var tagNames []string
		for _, tag := range bookmark.Tags {
			tagNames = append(tagNames, tag.Name)
		}
		secondaryText := strings.Join(tagNames, ", ")
		if secondaryText == "" {
			secondaryText = "No tags"
		}

		t.bookmarkList.AddItem(title, secondaryText, 0, nil)
	}

	// Update title
	if tag != "" {
		t.bookmarkList.SetTitle(fmt.Sprintf(" Bookmarks - Tag: %s ", tag))
	} else {
		t.bookmarkList.SetTitle(" Bookmarks ")
	}

	t.setStatus(fmt.Sprintf("[green]Loaded %d bookmarks[white]", len(t.currentBookmarks)))
}

// loadTags loads and displays tags
func (t *TUI) loadTags(tagList *tview.List) {
	var err error

	// Clear list
	tagList.Clear()

	// Load tags
	t.currentTags, err = t.bookmarkService.GetAllTags()
	if err != nil {
		t.setStatus(fmt.Sprintf("[red]Failed to load tags: %v[white]", err))
		return
	}

	// Add "All" option
	tagList.AddItem("All", "Show all bookmarks", 0, nil)

	// Add tags to list
	for _, tag := range t.currentTags {
		tagList.AddItem(tag.Name, "", 0, nil)
	}
}

// search performs a search and updates the results list
func (t *TUI) search(query string, results *tview.List) {
	if query == "" {
		t.setStatus("[yellow]Please enter a search query[white]")
		return
	}

	// Clear results
	results.Clear()

	// Perform search
	var err error
	t.currentBookmarks, err = t.searchService.Search(query, 100)
	if err != nil {
		t.setStatus(fmt.Sprintf("[red]Search failed: %v[white]", err))
		return
	}

	// Add results to list
	for _, bookmark := range t.currentBookmarks {
		title := bookmark.Title
		if title == "" {
			title = bookmark.URL
		}

		// Create secondary text with tags
		var tagNames []string
		for _, tag := range bookmark.Tags {
			tagNames = append(tagNames, tag.Name)
		}
		secondaryText := strings.Join(tagNames, ", ")
		if secondaryText == "" {
			secondaryText = "No tags"
		}

		results.AddItem(title, secondaryText, 0, nil)
	}

	results.SetTitle(fmt.Sprintf(" Search Results for '%s' ", query))
	t.setStatus(fmt.Sprintf("[green]Found %d results[white]", len(t.currentBookmarks)))

  // clear the form
  searchInput := t.searchPage.GetItem(0).(*tview.InputField)
  searchInput.SetText("")
  t.app.SetFocus(results)
}

// addBookmark adds a new bookmark
func (t *TUI) addBookmark(url string, tags []string) {
	if url == "" {
		t.setStatus("[yellow]Please enter a URL[white]")
		return
	}

	// Add bookmark
	bookmark, err := t.bookmarkService.Add(url, tags)
	if err != nil {
		t.setStatus(fmt.Sprintf("[red]Failed to add bookmark: %v[white]", err))
		return
	}

	// Index bookmark
	err = t.searchService.IndexBookmark(bookmark)
	if err != nil {
		t.setStatus(fmt.Sprintf("[yellow]Bookmark added but indexing failed: %v[white]", err))
		return
	}

	t.setStatus(fmt.Sprintf("[green]Added bookmark: %s[white]", bookmark.Title))

	// Show bookmark
	t.viewBookmark(bookmark)
}

// viewBookmark displays the details of a bookmark
func (t *TUI) viewBookmark(bookmark *model.Bookmark) {
	// Get text views
	detailsView := t.viewBookmarkPage.GetItem(0).(*tview.TextView)
	contentView := t.viewBookmarkPage.GetItem(1).(*tview.TextView)

	// Clear views
	detailsView.Clear()
	contentView.Clear()

	// Format details
	detailsView.SetText(fmt.Sprintf(
		"[yellow]Title:[white] %s\n"+
			"[yellow]URL:[white] %s\n"+
			"[yellow]Created:[white] %s\n"+
			"[yellow]Tags:[white] %s\n\n"+
			"[yellow]Description:[white] %s",
		bookmark.Title,
		bookmark.URL,
		bookmark.CreatedAt.Format("2006-01-02 15:04:05"),
		t.formatTags(bookmark.Tags),
		bookmark.Description,
	))

	// Set content
	contentView.SetText(fmt.Sprintf(
		"[yellow]Summary:[white]\n%s\n\n"+
			"[yellow]Content:[white]\n%s",
		bookmark.Summary,
		bookmark.Content,
	))

	// Update title
	t.viewBookmarkPage.GetItem(0).(*tview.TextView).SetTitle(fmt.Sprintf(" Bookmark: %s ", bookmark.Title))

	// Show page
	t.showPage("viewBookmark")
}

// formatTags formats a list of tags
func (t *TUI) formatTags(tags []model.Tag) string {
	var tagNames []string
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}
	result := strings.Join(tagNames, ", ")
	if result == "" {
		result = "None"
	}
	return result
}

// setStatus updates the status bar
func (t *TUI) setStatus(status string) {
	t.statusBar.SetText(status)
}

// showPage shows a page
func (t *TUI) showPage(name string) {
	t.pages.SwitchToPage(name)
}

// Run starts the TUI
func (t *TUI) Run() error {
	return t.app.SetRoot(t.pages, true).EnableMouse(true).Run()
}
