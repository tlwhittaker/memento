package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tlwhittaker/memento/internal/api"
	"github.com/tlwhittaker/memento/internal/config"
	"github.com/tlwhittaker/memento/internal/models"
)

// Screen represents the current screen state.
type Screen int

const (
	ScreenList Screen = iota
	ScreenDetail
	ScreenCreate
	ScreenEdit
)

// EditorMode represents the current editor mode (for vim bindings).
type EditorMode int

const (
	ModeNormal EditorMode = iota
	ModeInsert
	ModeVisual
)

// Model is the root Bubbletea model.
type Model struct {
	apiClient     *api.Client
	settings      *config.Settings
	currentScreen Screen
	width         int
	height        int

	// Data
	memos         []models.Memo
	selectedMemo  *models.Memo
	nextPageToken string

	// List screen state
	listCursor       int
	listOffset       int
	confirmingDelete bool

	// Detail screen state
	detailScroll int

	// Create screen state
	createContent string
	createCursor  int

	// Edit screen state
	editContent    string
	editOriginal   string
	editCursor     int
	editingMemo    *models.Memo
	previousScreen Screen // Where to return after edit

	// Vim mode state
	editorMode    EditorMode
	visualStart   int    // Start of visual selection
	pendingAction rune   // Pending operator (d, c, y)
	pendingCount  int    // Numeric prefix
	yankBuffer    string // Yanked text for paste

	// Undo/redo history
	createHistory *EditHistory
	editHistory   *EditHistory

	// System clipboard
	clipboard *Clipboard

	// Search state
	searchActive  bool
	searchQuery   string
	searchCursor  int
	filteredMemos []models.Memo

	// Unsaved changes dialog
	showingUnsavedDialog bool

	// Error state
	err     error
	loading bool

	// Message to display (e.g., "Memo created successfully")
	statusMessage string
}

// NewModel creates a new root model.
func NewModel(client *api.Client, settings *config.Settings) Model {
	// Apply theme from settings
	ApplyTheme(settings)

	editorMode := ModeInsert
	if settings.IsVimMode() {
		editorMode = ModeNormal
	}

	return Model{
		apiClient:     client,
		settings:      settings,
		currentScreen: ScreenList,
		memos:         []models.Memo{},
		editorMode:    editorMode,
		createHistory: NewEditHistory(100),
		editHistory:   NewEditHistory(100),
		clipboard:     NewClipboard(),
	}
}

// Init initializes the model and loads initial data.
func (m Model) Init() tea.Cmd {
	return m.loadMemos()
}

// Message types
type (
	memosLoadedMsg struct {
		memos         []models.Memo
		nextPageToken string
	}
	memoCreatedMsg struct {
		memo models.Memo
	}
	memoUpdatedMsg struct {
		memo models.Memo
	}
	memoDeletedMsg struct {
		name string
	}
	memoPinnedMsg struct {
		memo models.Memo
	}
	memoArchivedMsg struct {
		memo models.Memo
	}
	memoVisibilityMsg struct {
		memo models.Memo
	}
	errMsg struct {
		err error
	}
	statusMsg string
)

// Commands
func (m Model) loadMemos() tea.Cmd {
	return func() tea.Msg {
		resp, err := m.apiClient.ListMemos(50, "")
		if err != nil {
			return errMsg{err: err}
		}
		return memosLoadedMsg{
			memos:         models.FromAPIList(resp.Memos),
			nextPageToken: resp.NextPageToken,
		}
	}
}

func (m Model) createMemo(content string) tea.Cmd {
	return func() tea.Msg {
		memo, err := m.apiClient.CreateMemo(content)
		if err != nil {
			return errMsg{err: err}
		}
		return memoCreatedMsg{memo: models.FromAPI(*memo)}
	}
}

func (m Model) updateMemo(name, content string) tea.Cmd {
	return func() tea.Msg {
		memo, err := m.apiClient.UpdateMemo(name, content)
		if err != nil {
			return errMsg{err: err}
		}
		return memoUpdatedMsg{memo: models.FromAPI(*memo)}
	}
}

func (m Model) deleteMemo(name string) tea.Cmd {
	return func() tea.Msg {
		err := m.apiClient.DeleteMemo(name)
		if err != nil {
			return errMsg{err: err}
		}
		return memoDeletedMsg{name: name}
	}
}

func (m Model) togglePin(name string, currentlyPinned bool) tea.Cmd {
	return func() tea.Msg {
		memo, err := m.apiClient.SetMemoPinned(name, !currentlyPinned)
		if err != nil {
			return errMsg{err: err}
		}
		return memoPinnedMsg{memo: models.FromAPI(*memo)}
	}
}

func (m Model) toggleArchive(name string, isArchived bool) tea.Cmd {
	return func() tea.Msg {
		var memo *api.Memo
		var err error
		if isArchived {
			memo, err = m.apiClient.UnarchiveMemo(name)
		} else {
			memo, err = m.apiClient.ArchiveMemo(name)
		}
		if err != nil {
			return errMsg{err: err}
		}
		return memoArchivedMsg{memo: models.FromAPI(*memo)}
	}
}

func (m Model) cycleVisibility(name string, currentVisibility string) tea.Cmd {
	return func() tea.Msg {
		// Cycle: PRIVATE -> PROTECTED -> PUBLIC -> PRIVATE
		newVisibility := api.VisibilityPrivate
		switch currentVisibility {
		case api.VisibilityPrivate:
			newVisibility = api.VisibilityProtected
		case api.VisibilityProtected:
			newVisibility = api.VisibilityPublic
		case api.VisibilityPublic:
			newVisibility = api.VisibilityPrivate
		}

		memo, err := m.apiClient.SetMemoVisibility(name, newVisibility)
		if err != nil {
			return errMsg{err: err}
		}
		return memoVisibilityMsg{memo: models.FromAPI(*memo)}
	}
}
