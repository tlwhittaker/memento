package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kujtimiihoxha/vimtea"
	"github.com/sahilm/fuzzy"
	"github.com/tlwhittaker/memento/internal/models"
)

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear status message on any key press
		m.statusMessage = ""
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update VimTea editor sizes
		if m.createEditor != nil {
			m.createEditor.SetSize(msg.Width-8, msg.Height-12)
		}
		if m.editEditor != nil {
			m.editEditor.SetSize(msg.Width-8, msg.Height-12)
		}
		return m, nil

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case memosLoadedMsg:
		m.loading = false
		m.memos = msg.memos
		m.nextPageToken = msg.nextPageToken
		m.err = nil
		return m, nil

	case memoCreatedMsg:
		m.loading = false
		// Prepend new memo to list
		m.memos = append([]models.Memo{msg.memo}, m.memos...)
		m.currentScreen = ScreenList
		m.statusMessage = "Memo created"
		m.createEditor = nil
		return m, nil

	case memoUpdatedMsg:
		m.loading = false
		// Update memo in list
		for i, memo := range m.memos {
			if memo.Name == msg.memo.Name {
				m.memos[i] = msg.memo
				break
			}
		}
		// Update selectedMemo if viewing detail
		if m.selectedMemo != nil && m.selectedMemo.Name == msg.memo.Name {
			m.selectedMemo = &msg.memo
		}
		// Return to previous screen
		m.currentScreen = m.previousScreen
		m.statusMessage = "Memo updated"
		m.editEditor = nil
		m.editOriginal = ""
		m.editingMemo = nil
		return m, nil

	case memoDeletedMsg:
		m.loading = false
		// Remove deleted memo from list
		for i, memo := range m.memos {
			if memo.Name == msg.name {
				m.memos = append(m.memos[:i], m.memos[i+1:]...)
				break
			}
		}
		m.confirmingDelete = false
		m.currentScreen = ScreenList
		m.selectedMemo = nil
		m.statusMessage = "Memo deleted"
		// Adjust cursor if needed
		if m.listCursor >= len(m.memos) && m.listCursor > 0 {
			m.listCursor--
		}
		return m, nil

	case memoPinnedMsg:
		m.loading = false
		// Update memo in list
		for i, memo := range m.memos {
			if memo.Name == msg.memo.Name {
				m.memos[i] = msg.memo
				break
			}
		}
		if m.selectedMemo != nil && m.selectedMemo.Name == msg.memo.Name {
			m.selectedMemo = &msg.memo
		}
		if msg.memo.Pinned {
			m.statusMessage = "Memo pinned"
		} else {
			m.statusMessage = "Memo unpinned"
		}
		return m, nil

	case memoArchivedMsg:
		m.loading = false
		// Update memo in list
		for i, memo := range m.memos {
			if memo.Name == msg.memo.Name {
				m.memos[i] = msg.memo
				break
			}
		}
		if m.selectedMemo != nil && m.selectedMemo.Name == msg.memo.Name {
			m.selectedMemo = &msg.memo
		}
		if msg.memo.IsArchived() {
			m.statusMessage = "Memo archived"
		} else {
			m.statusMessage = "Memo unarchived"
		}
		return m, nil

	case memoVisibilityMsg:
		m.loading = false
		// Update memo in list
		for i, memo := range m.memos {
			if memo.Name == msg.memo.Name {
				m.memos[i] = msg.memo
				break
			}
		}
		if m.selectedMemo != nil && m.selectedMemo.Name == msg.memo.Name {
			m.selectedMemo = &msg.memo
		}
		m.statusMessage = "Visibility: " + msg.memo.Visibility
		return m, nil

	case errMsg:
		m.loading = false
		m.err = msg.err
		m.confirmingDelete = false
		m.showingUnsavedDialog = false
		return m, nil

	case statusMsg:
		m.statusMessage = string(msg)
		return m, nil

	// VimTea editor messages
	case saveRequestedMsg:
		if m.currentScreen == ScreenCreate && m.createEditor != nil {
			content := m.createEditor.GetBuffer().Text()
			if content != "" {
				m.loading = true
				return m, m.createMemo(content)
			}
		} else if m.currentScreen == ScreenEdit && m.editEditor != nil {
			content := m.editEditor.GetBuffer().Text()
			if content != "" && m.editingMemo != nil {
				m.loading = true
				return m, m.updateMemo(m.editingMemo.Name, content)
			}
		}
		return m, nil

	case cancelRequestedMsg:
		if m.currentScreen == ScreenCreate && m.createEditor != nil {
			content := m.createEditor.GetBuffer().Text()
			if content != "" {
				m.showingUnsavedDialog = true
				return m, nil
			}
			m.currentScreen = ScreenList
			m.createEditor = nil
		} else if m.currentScreen == ScreenEdit && m.editEditor != nil {
			content := m.editEditor.GetBuffer().Text()
			if content != m.editOriginal {
				m.showingUnsavedDialog = true
				return m, nil
			}
			m.currentScreen = m.previousScreen
			m.editEditor = nil
			m.editOriginal = ""
			m.editingMemo = nil
		}
		return m, nil

	case saveAndQuitMsg:
		if m.currentScreen == ScreenCreate && m.createEditor != nil {
			content := m.createEditor.GetBuffer().Text()
			if content != "" {
				m.loading = true
				return m, m.createMemo(content)
			}
			// Empty content, just quit
			m.currentScreen = ScreenList
			m.createEditor = nil
		} else if m.currentScreen == ScreenEdit && m.editEditor != nil {
			content := m.editEditor.GetBuffer().Text()
			if content != "" && m.editingMemo != nil {
				m.loading = true
				return m, m.updateMemo(m.editingMemo.Name, content)
			}
			// Empty content, just quit
			m.currentScreen = m.previousScreen
			m.editEditor = nil
			m.editOriginal = ""
			m.editingMemo = nil
		}
		return m, nil

	// Forward VimTea internal messages to the active editor
	case vimtea.CommandMsg, vimtea.UndoRedoMsg, vimtea.EditorModeMsg:
		if m.currentScreen == ScreenCreate && m.createEditor != nil {
			newEditor, cmd := m.createEditor.Update(msg)
			m.createEditor = newEditor.(vimtea.Editor)
			return m, cmd
		} else if m.currentScreen == ScreenEdit && m.editEditor != nil {
			newEditor, cmd := m.editEditor.Update(msg)
			m.editEditor = newEditor.(vimtea.Editor)
			return m, cmd
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	// Handle template picker first
	if m.showingTemplatePicker {
		return m.handleTemplatePicker(msg)
	}

	// Handle unsaved changes dialog
	if m.showingUnsavedDialog {
		return m.handleUnsavedConfirmation(msg)
	}

	// Handle delete confirmation
	if m.confirmingDelete {
		return m.handleDeleteConfirmation(msg)
	}

	switch m.currentScreen {
	case ScreenList:
		return m.handleListKeys(msg)
	case ScreenDetail:
		return m.handleDetailKeys(msg)
	case ScreenCreate:
		return m.handleCreateKeys(msg)
	case ScreenEdit:
		return m.handleEditKeys(msg)
	case ScreenCalendar:
		return m.handleCalendarKeys(msg)
	}

	return m, nil
}

func (m Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle search mode
	if m.searchActive {
		return m.handleSearchKeys(msg)
	}

	// Get the list of memos to operate on (filtered or all)
	memoList := m.getDisplayMemos()

	// Check if we're in split pane mode and if inline calendar is visible
	inSplitPane := m.width >= SplitPaneMinWidth
	showInlineCal := m.shouldShowInlineCalendar()

	// Handle inline calendar navigation when focused
	if showInlineCal && m.inlineCalendarFocus && !m.splitFocusRight {
		return m.handleInlineCalendarKeys(msg)
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "left", "h":
		// In split pane mode, switch focus to left pane
		if inSplitPane && m.splitFocusRight {
			m.splitFocusRight = false
			return m, nil
		}

	case "right", "l":
		// In split pane mode, switch focus to right pane
		if inSplitPane && !m.splitFocusRight && len(memoList) > 0 {
			m.splitFocusRight = true
			m.inlineCalendarFocus = false
			m.previewScroll = 0
			return m, nil
		}

	case "up", "k":
		if inSplitPane && m.splitFocusRight {
			// Scroll preview pane up
			if m.previewScroll > 0 {
				m.previewScroll--
			}
		} else {
			// Move list cursor up
			if m.listCursor > 0 {
				m.listCursor--
				m.adjustListScroll(m.height - 9)
				m.previewScroll = 0 // Reset preview scroll when changing selection
			}
		}

	case "down", "j":
		if inSplitPane && m.splitFocusRight {
			// Scroll preview pane down
			m.previewScroll++
		} else {
			// Move list cursor down
			if m.listCursor < len(memoList)-1 {
				m.listCursor++
				m.adjustListScroll(m.height - 9)
				m.previewScroll = 0 // Reset preview scroll when changing selection
			}
		}

	case "enter":
		if len(memoList) > 0 && m.listCursor < len(memoList) {
			m.selectedMemo = &memoList[m.listCursor]
			m.currentScreen = ScreenDetail
			m.detailScroll = 0
		}

	case "n":
		m.clearSearch()
		// Check if templates exist
		if len(m.templates) > 0 {
			// Show template picker
			m.showingTemplatePicker = true
			m.templateCursor = 0
		} else {
			// No templates - go directly to create screen with VimTea editor
			m.createEditor = newVimTeaEditor("")
			m.createEditor.SetSize(m.width-8, m.height-12)
			m.currentScreen = ScreenCreate
		}

	case "e":
		if len(memoList) > 0 && m.listCursor < len(memoList) {
			memo := &memoList[m.listCursor]
			m.editingMemo = memo
			m.editOriginal = memo.Content
			m.editEditor = newVimTeaEditor(memo.Content)
			m.editEditor.SetSize(m.width-8, m.height-12)
			m.previousScreen = ScreenList
			m.currentScreen = ScreenEdit
		}

	case "d":
		if len(memoList) > 0 && m.listCursor < len(memoList) {
			m.selectedMemo = &memoList[m.listCursor]
			m.confirmingDelete = true
		}

	case "r":
		m.loading = true
		m.clearSearch()
		return m, m.loadMemos()

	case "/":
		m.searchActive = true
		m.searchQuery = ""
		m.searchCursor = 0
		m.filterMemos()

	case "p":
		// Toggle pin
		if len(memoList) > 0 && m.listCursor < len(memoList) {
			memo := memoList[m.listCursor]
			m.loading = true
			return m, m.togglePin(memo.Name, memo.Pinned)
		}

	case "a":
		// Toggle archive
		if len(memoList) > 0 && m.listCursor < len(memoList) {
			memo := memoList[m.listCursor]
			m.loading = true
			return m, m.toggleArchive(memo.Name, memo.IsArchived())
		}

	case "v":
		// Cycle visibility
		if len(memoList) > 0 && m.listCursor < len(memoList) {
			memo := memoList[m.listCursor]
			m.loading = true
			return m, m.cycleVisibility(memo.Name, memo.Visibility)
		}

	case "c":
		// Toggle between list and calendar focus, or open full calendar
		if showInlineCal {
			m.inlineCalendarFocus = true
		} else {
			m.currentScreen = ScreenCalendar
		}
	}

	return m, nil
}

func (m Model) handleInlineCalendarKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	daysInMonth := m.daysInCurrentMonth()
	memoList := m.getDisplayMemos()

	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "c", "esc":
		// Switch back to memo list
		m.inlineCalendarFocus = false

	case "right", "l":
		// Move to next day, or switch to preview pane at month end
		m.calendarDay++
		if m.calendarDay > daysInMonth {
			m.calendarDay = 1
			m.calendarMonth++
			if m.calendarMonth > 12 {
				m.calendarMonth = 1
				m.calendarYear++
			}
		}

	case "left", "h":
		// Move to previous day
		m.calendarDay--
		if m.calendarDay < 1 {
			m.calendarMonth--
			if m.calendarMonth < 1 {
				m.calendarMonth = 12
				m.calendarYear--
			}
			m.calendarDay = m.daysInCurrentMonth()
		}

	case "down", "j":
		// Move down one week
		m.calendarDay += 7
		if m.calendarDay > daysInMonth {
			m.calendarDay -= daysInMonth
			m.calendarMonth++
			if m.calendarMonth > 12 {
				m.calendarMonth = 1
				m.calendarYear++
			}
		}

	case "up", "k":
		// Move up one week
		m.calendarDay -= 7
		if m.calendarDay < 1 {
			m.calendarMonth--
			if m.calendarMonth < 1 {
				m.calendarMonth = 12
				m.calendarYear--
			}
			m.calendarDay += m.daysInCurrentMonth()
		}

	case "H":
		// Previous month
		m.calendarMonth--
		if m.calendarMonth < 1 {
			m.calendarMonth = 12
			m.calendarYear--
		}
		maxDay := m.daysInCurrentMonth()
		if m.calendarDay > maxDay {
			m.calendarDay = maxDay
		}

	case "L":
		// Next month
		m.calendarMonth++
		if m.calendarMonth > 12 {
			m.calendarMonth = 1
			m.calendarYear++
		}
		maxDay := m.daysInCurrentMonth()
		if m.calendarDay > maxDay {
			m.calendarDay = maxDay
		}

	case "enter":
		// View memos for selected day - jump to first memo on that day
		dayMemos := m.getMemosForDate(m.calendarYear, m.calendarMonth, m.calendarDay)
		if len(dayMemos) > 0 {
			// Find the memo in the main list and set cursor
			for i, memo := range memoList {
				if memo.Name == dayMemos[0].Name {
					m.listCursor = i
					m.adjustListScroll(m.height - 9)
					break
				}
			}
			m.inlineCalendarFocus = false
		}

	case "tab":
		// Switch to preview pane
		if len(memoList) > 0 {
			m.splitFocusRight = true
			m.inlineCalendarFocus = false
			m.previewScroll = 0
		}
	}

	return m, nil
}

func (m Model) handleCalendarKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	daysInMonth := m.daysInCurrentMonth()

	switch msg.String() {
	case "q", "esc":
		m.currentScreen = ScreenList

	case "h", "left":
		m.calendarDay--
		if m.calendarDay < 1 {
			m.calendarMonth--
			if m.calendarMonth < 1 {
				m.calendarMonth = 12
				m.calendarYear--
			}
			m.calendarDay = m.daysInCurrentMonth()
		}

	case "l", "right":
		m.calendarDay++
		if m.calendarDay > daysInMonth {
			m.calendarDay = 1
			m.calendarMonth++
			if m.calendarMonth > 12 {
				m.calendarMonth = 1
				m.calendarYear++
			}
		}

	case "k", "up":
		m.calendarDay -= 7
		if m.calendarDay < 1 {
			m.calendarMonth--
			if m.calendarMonth < 1 {
				m.calendarMonth = 12
				m.calendarYear--
			}
			m.calendarDay += m.daysInCurrentMonth()
		}

	case "j", "down":
		m.calendarDay += 7
		if m.calendarDay > daysInMonth {
			m.calendarDay -= daysInMonth
			m.calendarMonth++
			if m.calendarMonth > 12 {
				m.calendarMonth = 1
				m.calendarYear++
			}
		}

	case "H":
		m.calendarMonth--
		if m.calendarMonth < 1 {
			m.calendarMonth = 12
			m.calendarYear--
		}
		maxDay := m.daysInCurrentMonth()
		if m.calendarDay > maxDay {
			m.calendarDay = maxDay
		}

	case "L":
		m.calendarMonth++
		if m.calendarMonth > 12 {
			m.calendarMonth = 1
			m.calendarYear++
		}
		maxDay := m.daysInCurrentMonth()
		if m.calendarDay > maxDay {
			m.calendarDay = maxDay
		}

	case "enter":
		dayMemos := m.getMemosForDate(m.calendarYear, m.calendarMonth, m.calendarDay)
		if len(dayMemos) > 0 {
			m.selectedMemo = &dayMemos[0]
			m.currentScreen = ScreenDetail
			m.detailScroll = 0
		}

	case "n":
		// Check if templates exist
		if len(m.templates) > 0 {
			// Show template picker
			m.showingTemplatePicker = true
			m.templateCursor = 0
		} else {
			// No templates - go directly to create screen with VimTea editor
			m.createEditor = newVimTeaEditor("")
			m.createEditor.SetSize(m.width-8, m.height-12)
			m.currentScreen = ScreenCreate
		}
	}

	return m, nil
}

func (m Model) daysInCurrentMonth() int {
	return m.daysInMonth(m.calendarYear, m.calendarMonth)
}

func (m Model) daysInMonth(year, month int) int {
	return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.Local).Day()
}

func (m Model) handleSearchKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		m.clearSearch()

	case "enter":
		// Confirm search and exit search mode
		m.searchActive = false

	case "backspace":
		if m.searchCursor > 0 {
			m.searchQuery = m.searchQuery[:m.searchCursor-1] + m.searchQuery[m.searchCursor:]
			m.searchCursor--
			m.filterMemos()
		}

	case "delete":
		if m.searchCursor < len(m.searchQuery) {
			m.searchQuery = m.searchQuery[:m.searchCursor] + m.searchQuery[m.searchCursor+1:]
			m.filterMemos()
		}

	case "left":
		if m.searchCursor > 0 {
			m.searchCursor--
		}

	case "right":
		if m.searchCursor < len(m.searchQuery) {
			m.searchCursor++
		}

	case "up", "ctrl+p":
		if m.listCursor > 0 {
			m.listCursor--
			m.adjustListScroll(m.height - 9)
		}

	case "down", "ctrl+n":
		memoList := m.getDisplayMemos()
		if m.listCursor < len(memoList)-1 {
			m.listCursor++
			m.adjustListScroll(m.height - 9)
		}

	default:
		// Handle regular character input
		if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
			m.searchQuery = m.searchQuery[:m.searchCursor] + key + m.searchQuery[m.searchCursor:]
			m.searchCursor++
			m.filterMemos()
		}
	}

	return m, nil
}

func (m *Model) clearSearch() {
	m.searchActive = false
	m.searchQuery = ""
	m.searchCursor = 0
	m.filteredMemos = nil
	m.listCursor = 0
}

func (m *Model) adjustListScroll(listHeight int) {
	if m.listCursor < m.listOffset {
		m.listOffset = m.listCursor
	}
	if listHeight > 0 && m.listCursor >= m.listOffset+listHeight {
		m.listOffset = m.listCursor - listHeight + 1
	}
}

// memoSearcher implements fuzzy.Source for memo content.
type memoSearcher struct {
	memos []models.Memo
}

func (s memoSearcher) String(i int) string {
	return s.memos[i].Content
}

func (s memoSearcher) Len() int {
	return len(s.memos)
}

// filterMemos filters memos based on the search query using fuzzy matching.
func (m *Model) filterMemos() {
	if m.searchQuery == "" {
		m.filteredMemos = nil
		return
	}

	// Use fuzzy matching
	source := memoSearcher{memos: m.memos}
	matches := fuzzy.FindFrom(m.searchQuery, source)

	m.filteredMemos = make([]models.Memo, 0, len(matches))
	for _, match := range matches {
		m.filteredMemos = append(m.filteredMemos, m.memos[match.Index])
	}

	// Reset cursor if it's out of range
	if m.listCursor >= len(m.filteredMemos) {
		m.listCursor = 0
	}
}

// getDisplayMemos returns the list of memos to display (filtered or all).
func (m *Model) getDisplayMemos() []models.Memo {
	if m.searchQuery != "" && m.filteredMemos != nil {
		return m.filteredMemos
	}
	return m.memos
}

func (m Model) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.currentScreen = ScreenList
		m.selectedMemo = nil

	case "up", "k":
		if m.detailScroll > 0 {
			m.detailScroll--
		}

	case "down", "j":
		m.detailScroll++

	case "e":
		if m.selectedMemo != nil {
			m.editingMemo = m.selectedMemo
			m.editOriginal = m.selectedMemo.Content
			m.editEditor = newVimTeaEditor(m.selectedMemo.Content)
			m.editEditor.SetSize(m.width-8, m.height-12)
			m.previousScreen = ScreenDetail
			m.currentScreen = ScreenEdit
		}

	case "d":
		m.confirmingDelete = true

	case "p":
		// Toggle pin
		if m.selectedMemo != nil {
			m.loading = true
			return m, m.togglePin(m.selectedMemo.Name, m.selectedMemo.Pinned)
		}

	case "a":
		// Toggle archive
		if m.selectedMemo != nil {
			m.loading = true
			return m, m.toggleArchive(m.selectedMemo.Name, m.selectedMemo.IsArchived())
		}

	case "v":
		// Cycle visibility
		if m.selectedMemo != nil {
			m.loading = true
			return m, m.cycleVisibility(m.selectedMemo.Name, m.selectedMemo.Visibility)
		}
	}

	return m, nil
}

func (m Model) handleCreateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.createEditor == nil {
		return m, nil
	}

	// Forward message to VimTea editor
	newEditor, cmd := m.createEditor.Update(msg)
	m.createEditor = newEditor.(vimtea.Editor)
	return m, cmd
}

func (m Model) handleEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editEditor == nil {
		return m, nil
	}

	// Forward message to VimTea editor
	newEditor, cmd := m.editEditor.Update(msg)
	m.editEditor = newEditor.(vimtea.Editor)
	return m, cmd
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch m.currentScreen {
	case ScreenList:
		return m.handleListMouse(msg)
	case ScreenDetail:
		return m.handleDetailMouse(msg)
	}
	return m, nil
}

func (m Model) handleListMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Don't handle mouse during dialogs
	if m.confirmingDelete || m.showingUnsavedDialog {
		return m, nil
	}

	memoList := m.getDisplayMemos()

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		if m.listCursor > 0 {
			m.listCursor--
			(&m).adjustListScroll(m.height - 9)
		}
	case tea.MouseButtonWheelDown:
		if m.listCursor < len(memoList)-1 {
			m.listCursor++
			(&m).adjustListScroll(m.height - 9)
		}
	case tea.MouseButtonLeft:
		// Calculate which memo was clicked based on Y position
		// Account for header (1 line), search bar (1 line if active), status (1 line if present)
		contentStartY := 3 // Header + border
		if m.searchActive || m.searchQuery != "" {
			contentStartY++
		}
		if m.err != nil || m.statusMessage != "" {
			contentStartY++
		}

		clickedIndex := msg.Y - contentStartY + m.listOffset
		if clickedIndex >= 0 && clickedIndex < len(memoList) {
			if m.listCursor == clickedIndex {
				m.selectedMemo = &memoList[clickedIndex]
				m.currentScreen = ScreenDetail
				m.detailScroll = 0
			} else {
				m.listCursor = clickedIndex
				(&m).adjustListScroll(m.height - 9)
			}
		}
	}

	return m, nil
}

func (m Model) handleDetailMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.confirmingDelete {
		return m, nil
	}

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		if m.detailScroll > 0 {
			m.detailScroll--
		}
	case tea.MouseButtonWheelDown:
		m.detailScroll++
	case tea.MouseButtonRight:
		// Right-click to go back
		m.currentScreen = ScreenList
		m.selectedMemo = nil
	}

	return m, nil
}

func (m Model) handleDeleteConfirmation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.selectedMemo != nil {
			m.loading = true
			return m, m.deleteMemo(m.selectedMemo.Name)
		}
		m.confirmingDelete = false

	case "n", "N", "esc":
		m.confirmingDelete = false
	}

	return m, nil
}

func (m Model) handleUnsavedConfirmation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Discard changes
		m.showingUnsavedDialog = false
		if m.currentScreen == ScreenCreate {
			m.currentScreen = ScreenList
			m.createEditor = nil
		} else if m.currentScreen == ScreenEdit {
			m.currentScreen = m.previousScreen
			m.editEditor = nil
			m.editOriginal = ""
			m.editingMemo = nil
		}

	case "n", "N", "esc":
		// Keep editing
		m.showingUnsavedDialog = false
	}

	return m, nil
}

func (m Model) handleTemplatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Skip template selection, start with blank note
		m.showingTemplatePicker = false
		m.createEditor = newVimTeaEditor("")
		m.createEditor.SetSize(m.width-8, m.height-12)
		m.currentScreen = ScreenCreate

	case "up", "k":
		if m.templateCursor > 0 {
			m.templateCursor--
		}

	case "down", "j":
		if m.templateCursor < len(m.templates)-1 {
			m.templateCursor++
		}

	case "enter":
		// Apply selected template
		m.showingTemplatePicker = false

		content := ""
		if m.templateCursor < len(m.templates) {
			content = m.templates[m.templateCursor].Content
		}

		m.createEditor = newVimTeaEditor(content)
		m.createEditor.SetSize(m.width-8, m.height-12)
		m.currentScreen = ScreenCreate
	}

	return m, nil
}
