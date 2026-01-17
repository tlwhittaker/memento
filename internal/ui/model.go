package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kujtimiihoxha/vimtea"

	"github.com/tlwhittaker/memento/internal/api"
	"github.com/tlwhittaker/memento/internal/config"
	"github.com/tlwhittaker/memento/internal/export"
	"github.com/tlwhittaker/memento/internal/models"
)

// Screen represents the current screen state.
type Screen int

const (
	ScreenList Screen = iota
	ScreenDetail
	ScreenCreate
	ScreenEdit
	ScreenCalendar
	ScreenTags
)

const (
	SplitPaneMinWidth = 120
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

	// Split pane state
	splitFocusRight bool
	previewScroll   int

	// Calendar state
	calendarYear        int
	calendarMonth       int
	calendarDay         int
	inlineCalendarFocus bool // When true, focus is on inline calendar in left pane

	// VimTea editors for create and edit screens
	createEditor vimtea.Editor
	editEditor   vimtea.Editor

	// Simple textarea editors (alternative to vim mode)
	simpleCreateEditor textarea.Model
	simpleEditEditor   textarea.Model

	// Editor mode: "vim" or "simple"
	editorMode string

	// Edit screen state
	editOriginal   string // Original content for unsaved changes check
	editingMemo    *models.Memo
	previousScreen Screen // Where to return after edit

	// System clipboard (for integration)
	clipboard *Clipboard

	// Search state
	searchActive  bool
	searchQuery   string
	searchCursor  int
	filteredMemos []models.Memo

	// Unsaved changes dialog
	showingUnsavedDialog bool

	// Template picker state
	showingTemplatePicker bool
	templates             []config.Template
	templateCursor        int

	// Error state
	err     error
	loading bool

	// Message to display (e.g., "Memo created successfully")
	statusMessage string

	// Help overlay
	showingHelp bool

	// Command palette
	showingCommandPalette bool
	commandPaletteQuery   string
	commandPaletteCursor  int
	commands              []Command
	filteredCommands      []Command

	// Tags browser
	allTags         []TagInfo
	tagCursor       int
	tagScrollOffset int

	// Multi-select
	selectedMemos map[string]bool
	selectionMode bool

	// Markdown rendering toggle
	detailRenderMarkdown bool

	// View density (compact, comfortable, expanded)
	viewDensity string

	// Quick navigation pending key (for gg, gt, gc, gl)
	pendingKey string

	// Pagination
	totalMemoCount   int
	hasMorePages     bool
	loadingMoreMemos bool
}

// NewModel creates a new root model.
func NewModel(client *api.Client, settings *config.Settings) Model {
	// Apply theme from settings
	ApplyTheme(settings)

	now := time.Now()

	// Load templates (ignore errors - empty list is fine)
	templates, _ := config.LoadTemplates()

	// Determine editor mode from settings
	editorMode := "vim"
	if settings.Editor.Mode == "simple" {
		editorMode = "simple"
	}

	return Model{
		apiClient:     client,
		settings:      settings,
		currentScreen: ScreenList,
		memos:         []models.Memo{},
		clipboard:     NewClipboard(),
		calendarYear:  now.Year(),
		calendarMonth: int(now.Month()),
		calendarDay:   now.Day(),
		templates:     templates,
		commands:      GetCommands(),
		selectedMemos: make(map[string]bool),
		viewDensity:   "comfortable",
		editorMode:    editorMode,
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

	// VimTea editor messages
	saveRequestedMsg   struct{}
	cancelRequestedMsg struct{}
	saveAndQuitMsg     struct{}
)

// newVimTeaEditor creates a configured VimTea editor for memo editing.
func newVimTeaEditor(content string) vimtea.Editor {
	editor := vimtea.NewEditor(
		vimtea.WithContent(content),
		vimtea.WithEnableStatusBar(false), // We use memento's status bar
		vimtea.WithRelativeNumbers(false),
	)

	// Add ctrl+s binding for save in all modes
	editor.AddBinding(vimtea.KeyBinding{
		Key:         "ctrl+s",
		Mode:        vimtea.ModeNormal,
		Description: "Save memo",
		Handler: func(b vimtea.Buffer) tea.Cmd {
			return func() tea.Msg { return saveRequestedMsg{} }
		},
	})
	editor.AddBinding(vimtea.KeyBinding{
		Key:         "ctrl+s",
		Mode:        vimtea.ModeInsert,
		Description: "Save memo",
		Handler: func(b vimtea.Buffer) tea.Cmd {
			return func() tea.Msg { return saveRequestedMsg{} }
		},
	})
	editor.AddBinding(vimtea.KeyBinding{
		Key:         "ctrl+s",
		Mode:        vimtea.ModeVisual,
		Description: "Save memo",
		Handler: func(b vimtea.Buffer) tea.Cmd {
			return func() tea.Msg { return saveRequestedMsg{} }
		},
	})

	// Add :w, :q, and :wq commands
	editor.AddCommand("w", func(b vimtea.Buffer, args []string) tea.Cmd {
		return func() tea.Msg { return saveRequestedMsg{} }
	})
	editor.AddCommand("q", func(b vimtea.Buffer, args []string) tea.Cmd {
		return func() tea.Msg { return cancelRequestedMsg{} }
	})
	editor.AddCommand("wq", func(b vimtea.Buffer, args []string) tea.Cmd {
		return func() tea.Msg { return saveAndQuitMsg{} }
	})

	return editor
}

// newSimpleEditor creates a configured textarea for simple text editing.
func newSimpleEditor(content string, width, height int) textarea.Model {
	ta := textarea.New()
	ta.SetValue(content)
	ta.SetWidth(width)
	ta.SetHeight(height)
	ta.Focus()
	ta.ShowLineNumbers = false
	ta.CharLimit = 0 // No limit
	ta.Placeholder = "Start typing..."
	return ta
}

// initCreateEditor initializes the create editor based on current mode.
func (m *Model) initCreateEditor(content string) {
	if m.editorMode == "vim" {
		m.createEditor = newVimTeaEditor(content)
		m.createEditor.SetSize(m.width-8, m.height-12)
	} else {
		m.simpleCreateEditor = newSimpleEditor(content, m.width-8, m.height-12)
	}
}

// initEditEditor initializes the edit editor based on current mode.
func (m *Model) initEditEditor(content string) {
	if m.editorMode == "vim" {
		m.editEditor = newVimTeaEditor(content)
		m.editEditor.SetSize(m.width-8, m.height-12)
	} else {
		m.simpleEditEditor = newSimpleEditor(content, m.width-8, m.height-12)
	}
}

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

// applyAdvancedFilter applies the advanced filter to memos.
func (m *Model) applyAdvancedFilter() {
	if m.searchQuery == "" {
		m.filteredMemos = nil
		return
	}

	filter := ParseFilter(m.searchQuery)
	m.filteredMemos = ApplyFilter(m.memos, filter)

	// Reset cursor if out of range
	if m.listCursor >= len(m.filteredMemos) {
		m.listCursor = 0
	}
}

// updateTags refreshes the tag list from current memos.
func (m *Model) updateTags() {
	m.allTags = ExtractTags(m.memos)
}

// exportCurrentMemo exports the currently selected memo.
func (m Model) exportCurrentMemo(format string) tea.Cmd {
	return func() tea.Msg {
		memoList := m.getDisplayMemos()
		if len(memoList) == 0 || m.listCursor >= len(memoList) {
			return statusMsg("No memo to export")
		}

		memo := memoList[m.listCursor]
		var expFormat export.ExportFormat
		switch format {
		case "markdown":
			expFormat = export.FormatMarkdown
		case "json":
			expFormat = export.FormatJSON
		default:
			expFormat = export.FormatText
		}

		filepath, err := export.ExportMemo(memo, expFormat, "")
		if err != nil {
			return statusMsg("Export failed: " + err.Error())
		}
		return statusMsg("Exported to: " + filepath)
	}
}

// exportAllMemos exports all memos.
func (m Model) exportAllMemos() tea.Cmd {
	return func() tea.Msg {
		if len(m.memos) == 0 {
			return statusMsg("No memos to export")
		}

		dir, err := export.ExportMemos(m.memos, export.FormatMarkdown)
		if err != nil {
			return statusMsg("Export failed: " + err.Error())
		}
		return statusMsg("Exported to: " + dir)
	}
}

// loadMoreMemos loads the next page of memos.
func (m Model) loadMoreMemos() tea.Cmd {
	if m.nextPageToken == "" {
		return nil
	}
	return func() tea.Msg {
		resp, err := m.apiClient.ListMemos(50, m.nextPageToken)
		if err != nil {
			return errMsg{err: err}
		}
		return memosAppendedMsg{
			memos:         models.FromAPIList(resp.Memos),
			nextPageToken: resp.NextPageToken,
		}
	}
}

// Message type for appending memos
type memosAppendedMsg struct {
	memos         []models.Memo
	nextPageToken string
}

// clearSelection clears all selected memos.
func (m *Model) clearSelection() {
	m.selectedMemos = make(map[string]bool)
	m.selectionMode = false
}

// toggleSelection toggles selection for a memo.
func (m *Model) toggleSelection(name string) {
	if m.selectedMemos[name] {
		delete(m.selectedMemos, name)
	} else {
		m.selectedMemos[name] = true
	}
}

// selectedCount returns the number of selected memos.
func (m *Model) selectedCount() int {
	return len(m.selectedMemos)
}
