package ui

import (
	tea "github.com/charmbracelet/bubbletea"
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
		m.createContent = ""
		m.createCursor = 0
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
		m.editContent = ""
		m.editOriginal = ""
		m.editCursor = 0
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
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	// Handle unsaved changes dialog first
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

	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "up", "k":
		if m.listCursor > 0 {
			m.listCursor--
			m.adjustListScroll(m.height - 9)
		}

	case "down", "j":
		if m.listCursor < len(memoList)-1 {
			m.listCursor++
			m.adjustListScroll(m.height - 9)
		}

	case "enter":
		if len(memoList) > 0 && m.listCursor < len(memoList) {
			m.selectedMemo = &memoList[m.listCursor]
			m.currentScreen = ScreenDetail
			m.detailScroll = 0
		}

	case "n":
		m.currentScreen = ScreenCreate
		m.createContent = ""
		m.createCursor = 0
		m.clearSearch()
		// Reset editor state for vim mode
		if m.settings.IsVimMode() {
			m.editorMode = ModeNormal
		}
		m.createHistory = NewEditHistory(100)
		m.pendingAction = 0
		m.visualStart = 0

	case "e":
		if len(memoList) > 0 && m.listCursor < len(memoList) {
			memo := &memoList[m.listCursor]
			m.editingMemo = memo
			m.editContent = memo.Content
			m.editOriginal = memo.Content
			m.editCursor = len(memo.Content)
			m.previousScreen = ScreenList
			m.currentScreen = ScreenEdit
			// Reset editor state for vim mode
			if m.settings.IsVimMode() {
				m.editorMode = ModeNormal
			}
			m.editHistory = NewEditHistory(100)
			m.pendingAction = 0
			m.visualStart = 0
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
	}

	return m, nil
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
			m.editContent = m.selectedMemo.Content
			m.editOriginal = m.selectedMemo.Content
			m.editCursor = len(m.selectedMemo.Content)
			m.previousScreen = ScreenDetail
			m.currentScreen = ScreenEdit
			// Reset editor state for vim mode
			if m.settings.IsVimMode() {
				m.editorMode = ModeNormal
			}
			m.editHistory = NewEditHistory(100)
			m.pendingAction = 0
			m.visualStart = 0
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
	// Use vim bindings if enabled
	if m.settings.IsVimMode() {
		return m.handleVimCreateKeys(msg)
	}
	return m.handleNormalCreateKeys(msg)
}

func (m Model) handleNormalCreateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		// Check for unsaved changes
		if m.createContent != "" {
			m.showingUnsavedDialog = true
			return m, nil
		}
		m.currentScreen = ScreenList
		m.createContent = ""
		m.createCursor = 0

	case "ctrl+s":
		if m.createContent != "" {
			m.loading = true
			return m, m.createMemo(m.createContent)
		}

	case "enter":
		m.createContent = m.createContent[:m.createCursor] + "\n" + m.createContent[m.createCursor:]
		m.createCursor++

	case "backspace":
		if m.createCursor > 0 {
			m.createContent = m.createContent[:m.createCursor-1] + m.createContent[m.createCursor:]
			m.createCursor--
		}

	case "delete":
		if m.createCursor < len(m.createContent) {
			m.createContent = m.createContent[:m.createCursor] + m.createContent[m.createCursor+1:]
		}

	case "left":
		if m.createCursor > 0 {
			m.createCursor--
		}

	case "right":
		if m.createCursor < len(m.createContent) {
			m.createCursor++
		}

	case "up":
		m.createCursor = MoveUp(m.createContent, m.createCursor)

	case "down":
		m.createCursor = MoveDown(m.createContent, m.createCursor)

	case "home":
		m.createCursor = LineStart(m.createContent, m.createCursor)

	case "end":
		m.createCursor = LineEnd(m.createContent, m.createCursor)

	case "ctrl+a":
		// Select all / go to start
		m.createCursor = 0

	case "ctrl+e":
		// Go to end
		m.createCursor = len(m.createContent)

	case "ctrl+left":
		// Move word backward
		m.createCursor = WordBackward(m.createContent, m.createCursor)

	case "ctrl+right":
		// Move word forward
		m.createCursor = WordForward(m.createContent, m.createCursor)

	case "ctrl+backspace":
		// Delete word backward
		if m.createCursor > 0 {
			newPos := WordBackward(m.createContent, m.createCursor)
			m.createContent = m.createContent[:newPos] + m.createContent[m.createCursor:]
			m.createCursor = newPos
		}

	case "ctrl+delete":
		// Delete word forward
		if m.createCursor < len(m.createContent) {
			endPos := WordForward(m.createContent, m.createCursor)
			m.createContent = m.createContent[:m.createCursor] + m.createContent[endPos:]
		}

	case "ctrl+v":
		// Paste from clipboard
		if text, err := m.clipboard.Paste(); err == nil && text != "" {
			m.createContent = m.createContent[:m.createCursor] + text + m.createContent[m.createCursor:]
			m.createCursor += len(text)
		}

	case "ctrl+z":
		// Undo
		if content, cursor, ok := m.createHistory.Undo(); ok {
			m.createContent = content
			m.createCursor = cursor
		}

	case "ctrl+y", "ctrl+shift+z":
		// Redo
		if content, cursor, ok := m.createHistory.Redo(); ok {
			m.createContent = content
			m.createCursor = cursor
		}

	case "tab":
		// Insert tab as spaces
		spaces := "    "
		m.createContent = m.createContent[:m.createCursor] + spaces + m.createContent[m.createCursor:]
		m.createCursor += len(spaces)

	default:
		// Handle regular character input (printable ASCII and extended)
		if len(key) == 1 && key[0] >= 32 {
			m.createContent = m.createContent[:m.createCursor] + key + m.createContent[m.createCursor:]
			m.createCursor++
		}
	}

	return m, nil
}

func (m Model) handleVimCreateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Always handle ctrl+s for save
	if key == "ctrl+s" {
		if m.createContent != "" {
			m.loading = true
			return m, m.createMemo(m.createContent)
		}
		return m, nil
	}

	switch m.editorMode {
	case ModeNormal:
		return m.handleVimNormalMode(msg, true)
	case ModeInsert:
		return m.handleVimInsertMode(msg, true)
	case ModeVisual:
		return m.handleVimVisualMode(msg, true)
	}

	return m, nil
}

func (m Model) handleVimNormalMode(msg tea.KeyMsg, isCreate bool) (tea.Model, tea.Cmd) {
	key := msg.String()
	content := m.getEditorContent(isCreate)
	cursor := m.getEditorCursor(isCreate)

	// Handle pending operators (d, c, y followed by motion)
	if m.pendingAction != 0 {
		return m.handlePendingOperator(msg, isCreate)
	}

	switch key {
	// Exit to list
	case "esc":
		if isCreate && content == "" {
			m.currentScreen = ScreenList
			m.createContent = ""
			m.createCursor = 0
			return m, nil
		}
		// In normal mode, esc checks for unsaved changes
		if isCreate {
			if content != "" {
				m.showingUnsavedDialog = true
			} else {
				m.currentScreen = ScreenList
			}
		} else {
			if content != m.editOriginal {
				m.showingUnsavedDialog = true
			} else {
				m.currentScreen = m.previousScreen
				m.editContent = ""
				m.editOriginal = ""
				m.editCursor = 0
				m.editingMemo = nil
			}
		}

	// Mode switching
	case "i":
		m.editorMode = ModeInsert
	case "I":
		cursor = FirstNonBlank(content, cursor)
		m.editorMode = ModeInsert
	case "a":
		if cursor < len(content) {
			cursor++
		}
		m.editorMode = ModeInsert
	case "A":
		cursor = LineEnd(content, cursor)
		m.editorMode = ModeInsert
	case "o":
		content, cursor = InsertNewLineBelow(content, cursor)
		m.editorMode = ModeInsert
	case "O":
		content, cursor = InsertNewLineAbove(content, cursor)
		m.editorMode = ModeInsert
	case "v":
		m.editorMode = ModeVisual
		m.visualStart = cursor

	// Movement
	case "h", "left":
		if cursor > 0 {
			cursor--
		}
	case "l", "right":
		if cursor < len(content) {
			cursor++
		}
	case "j", "down":
		cursor = MoveDown(content, cursor)
	case "k", "up":
		cursor = MoveUp(content, cursor)
	case "w":
		cursor = WordForward(content, cursor)
	case "b":
		cursor = WordBackward(content, cursor)
	case "e":
		cursor = WordEnd(content, cursor)
	case "0":
		cursor = LineStart(content, cursor)
	case "^":
		cursor = FirstNonBlank(content, cursor)
	case "$":
		cursor = LineEnd(content, cursor)
	case "g":
		// gg goes to start (need to handle double key)
		m.pendingAction = 'g'
	case "G":
		cursor = DocumentEnd(content)

	// Delete operations
	case "x":
		if cursor < len(content) {
			m.pushHistory(isCreate, content, cursor)
			m.yankBuffer = string(content[cursor])
			content = content[:cursor] + content[cursor+1:]
		}
	case "d":
		m.pendingAction = 'd'
	case "D":
		m.pushHistory(isCreate, content, cursor)
		m.yankBuffer = content[cursor:LineEnd(content, cursor)]
		content, cursor = DeleteToLineEnd(content, cursor)
	case "c":
		m.pendingAction = 'c'
	case "C":
		m.pushHistory(isCreate, content, cursor)
		m.yankBuffer = content[cursor:LineEnd(content, cursor)]
		content, cursor = ChangeToLineEnd(content, cursor)
		m.editorMode = ModeInsert

	// Yank operations
	case "y":
		m.pendingAction = 'y'

	// Paste
	case "p":
		if m.yankBuffer != "" {
			m.pushHistory(isCreate, content, cursor)
			content, cursor = PasteAfter(content, cursor, m.yankBuffer)
		}
	case "P":
		if m.yankBuffer != "" {
			m.pushHistory(isCreate, content, cursor)
			content, cursor = PasteBefore(content, cursor, m.yankBuffer)
		}

	// Undo/Redo
	case "u":
		if newContent, newCursor, ok := m.undo(isCreate); ok {
			content = newContent
			cursor = newCursor
		}
	case "ctrl+r":
		if newContent, newCursor, ok := m.redo(isCreate); ok {
			content = newContent
			cursor = newCursor
		}
	}

	// Apply changes back to model
	m = m.setEditorContent(isCreate, content)
	m = m.setEditorCursor(isCreate, cursor)
	return m, nil
}

// pushHistory saves the current state to history.
func (m *Model) pushHistory(isCreate bool, content string, cursor int) {
	if isCreate {
		m.createHistory.Push(content, cursor)
	} else {
		m.editHistory.Push(content, cursor)
	}
}

// undo returns the previous state from history.
func (m *Model) undo(isCreate bool) (string, int, bool) {
	if isCreate {
		return m.createHistory.Undo()
	}
	return m.editHistory.Undo()
}

// redo returns the next state from history.
func (m *Model) redo(isCreate bool) (string, int, bool) {
	if isCreate {
		return m.createHistory.Redo()
	}
	return m.editHistory.Redo()
}

// getEditorContent returns the content for the current editor (create or edit).
func (m Model) getEditorContent(isCreate bool) string {
	if isCreate {
		return m.createContent
	}
	return m.editContent
}

// setEditorContent sets the content for the current editor and returns the updated model.
func (m Model) setEditorContent(isCreate bool, content string) Model {
	if isCreate {
		m.createContent = content
	} else {
		m.editContent = content
	}
	return m
}

// getEditorCursor returns the cursor position for the current editor.
func (m Model) getEditorCursor(isCreate bool) int {
	if isCreate {
		return m.createCursor
	}
	return m.editCursor
}

// setEditorCursor sets the cursor position for the current editor and returns the updated model.
func (m Model) setEditorCursor(isCreate bool, cursor int) Model {
	if isCreate {
		m.createCursor = cursor
	} else {
		m.editCursor = cursor
	}
	return m
}

func (m Model) handlePendingOperator(msg tea.KeyMsg, isCreate bool) (tea.Model, tea.Cmd) {
	key := msg.String()
	action := m.pendingAction
	m.pendingAction = 0
	content := m.getEditorContent(isCreate)
	cursor := m.getEditorCursor(isCreate)

	switch action {
	case 'g':
		if key == "g" {
			cursor = DocumentStart()
		}
	case 'd':
		m.pushHistory(isCreate, content, cursor)
		switch key {
		case "d":
			m.yankBuffer = YankLine(content, cursor)
			content, cursor = DeleteLine(content, cursor)
		case "w":
			m.yankBuffer = YankWord(content, cursor)
			content, cursor = DeleteWord(content, cursor)
		case "$":
			m.yankBuffer = content[cursor:LineEnd(content, cursor)]
			content, cursor = DeleteToLineEnd(content, cursor)
		}
	case 'c':
		m.pushHistory(isCreate, content, cursor)
		switch key {
		case "c":
			m.yankBuffer = GetCurrentLine(content, cursor)
			content, cursor = DeleteLine(content, cursor)
			// Insert newline at cursor position for cc
			content = content[:cursor] + "\n" + content[cursor:]
			m.editorMode = ModeInsert
		case "w":
			m.yankBuffer = YankWord(content, cursor)
			content, cursor = ChangeWord(content, cursor)
			m.editorMode = ModeInsert
		case "$":
			m.yankBuffer = content[cursor:LineEnd(content, cursor)]
			content, cursor = ChangeToLineEnd(content, cursor)
			m.editorMode = ModeInsert
		}
	case 'y':
		switch key {
		case "y":
			m.yankBuffer = YankLine(content, cursor)
			m.clipboard.Copy(m.yankBuffer)
		case "w":
			m.yankBuffer = YankWord(content, cursor)
			m.clipboard.Copy(m.yankBuffer)
		case "$":
			m.yankBuffer = content[cursor:LineEnd(content, cursor)]
			m.clipboard.Copy(m.yankBuffer)
		}
	}

	// Apply changes back to model
	m = m.setEditorContent(isCreate, content)
	m = m.setEditorCursor(isCreate, cursor)
	return m, nil
}

func (m Model) handleVimInsertMode(msg tea.KeyMsg, isCreate bool) (tea.Model, tea.Cmd) {
	key := msg.String()
	content := m.getEditorContent(isCreate)
	cursor := m.getEditorCursor(isCreate)

	switch key {
	case "esc":
		// Save state before exiting insert mode
		m.pushHistory(isCreate, content, cursor)
		m.editorMode = ModeNormal
		if cursor > 0 {
			cursor--
		}

	case "enter":
		content = content[:cursor] + "\n" + content[cursor:]
		cursor++

	case "backspace":
		if cursor > 0 {
			content = content[:cursor-1] + content[cursor:]
			cursor--
		}

	case "delete":
		if cursor < len(content) {
			content = content[:cursor] + content[cursor+1:]
		}

	case "left":
		if cursor > 0 {
			cursor--
		}

	case "right":
		if cursor < len(content) {
			cursor++
		}

	case "up":
		cursor = MoveUp(content, cursor)

	case "down":
		cursor = MoveDown(content, cursor)

	case "home":
		cursor = LineStart(content, cursor)

	case "end":
		cursor = LineEnd(content, cursor)

	case "ctrl+v":
		// Paste from system clipboard
		if text, err := m.clipboard.Paste(); err == nil && text != "" {
			content = content[:cursor] + text + content[cursor:]
			cursor += len(text)
		}

	case "ctrl+left":
		// Move word backward
		cursor = WordBackward(content, cursor)

	case "ctrl+right":
		// Move word forward
		cursor = WordForward(content, cursor)

	case "ctrl+backspace":
		// Delete word backward
		if cursor > 0 {
			newPos := WordBackward(content, cursor)
			content = content[:newPos] + content[cursor:]
			cursor = newPos
		}

	case "ctrl+delete":
		// Delete word forward
		if cursor < len(content) {
			endPos := WordForward(content, cursor)
			content = content[:cursor] + content[endPos:]
		}

	case "tab":
		// Insert tab as spaces
		spaces := "    "
		content = content[:cursor] + spaces + content[cursor:]
		cursor += len(spaces)

	default:
		// Handle regular character input (printable ASCII and extended)
		if len(key) == 1 && key[0] >= 32 {
			content = content[:cursor] + key + content[cursor:]
			cursor++
		}
	}

	// Apply changes back to model
	m = m.setEditorContent(isCreate, content)
	m = m.setEditorCursor(isCreate, cursor)
	return m, nil
}

func (m Model) handleVimVisualMode(msg tea.KeyMsg, isCreate bool) (tea.Model, tea.Cmd) {
	key := msg.String()
	content := m.getEditorContent(isCreate)
	cursor := m.getEditorCursor(isCreate)

	switch key {
	case "esc":
		m.editorMode = ModeNormal

	// Movement (extends selection)
	case "h", "left":
		if cursor > 0 {
			cursor--
		}
	case "l", "right":
		if cursor < len(content) {
			cursor++
		}
	case "j", "down":
		cursor = MoveDown(content, cursor)
	case "k", "up":
		cursor = MoveUp(content, cursor)
	case "w":
		cursor = WordForward(content, cursor)
	case "b":
		cursor = WordBackward(content, cursor)
	case "e":
		cursor = WordEnd(content, cursor)
	case "0":
		cursor = LineStart(content, cursor)
	case "$":
		cursor = LineEnd(content, cursor)

	// Operations on selection
	case "d", "x":
		m.pushHistory(isCreate, content, cursor)
		m.yankBuffer = GetSelectedText(content, m.visualStart, cursor)
		m.clipboard.Copy(m.yankBuffer)
		content, cursor = DeleteSelection(content, m.visualStart, cursor)
		m.editorMode = ModeNormal
	case "y":
		m.yankBuffer = GetSelectedText(content, m.visualStart, cursor)
		m.clipboard.Copy(m.yankBuffer)
		m.editorMode = ModeNormal
	case "c":
		m.pushHistory(isCreate, content, cursor)
		m.yankBuffer = GetSelectedText(content, m.visualStart, cursor)
		m.clipboard.Copy(m.yankBuffer)
		content, cursor = DeleteSelection(content, m.visualStart, cursor)
		m.editorMode = ModeInsert
	}

	// Apply changes back to model
	m = m.setEditorContent(isCreate, content)
	m = m.setEditorCursor(isCreate, cursor)
	return m, nil
}

func (m Model) handleEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Use vim bindings if enabled
	if m.settings.IsVimMode() {
		return m.handleVimEditKeys(msg)
	}
	return m.handleNormalEditKeys(msg)
}

func (m Model) handleNormalEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Check for unsaved changes
		if m.editContent != m.editOriginal {
			m.showingUnsavedDialog = true
			return m, nil
		}
		m.currentScreen = m.previousScreen
		m.editContent = ""
		m.editOriginal = ""
		m.editCursor = 0
		m.editingMemo = nil

	case "ctrl+s":
		if m.editContent != "" && m.editingMemo != nil {
			m.loading = true
			return m, m.updateMemo(m.editingMemo.Name, m.editContent)
		}

	case "enter":
		m.editContent = m.editContent[:m.editCursor] + "\n" + m.editContent[m.editCursor:]
		m.editCursor++

	case "backspace":
		if m.editCursor > 0 {
			m.editContent = m.editContent[:m.editCursor-1] + m.editContent[m.editCursor:]
			m.editCursor--
		}

	case "delete":
		if m.editCursor < len(m.editContent) {
			m.editContent = m.editContent[:m.editCursor] + m.editContent[m.editCursor+1:]
		}

	case "left":
		if m.editCursor > 0 {
			m.editCursor--
		}

	case "right":
		if m.editCursor < len(m.editContent) {
			m.editCursor++
		}

	case "up":
		m.editCursor = MoveUp(m.editContent, m.editCursor)

	case "down":
		m.editCursor = MoveDown(m.editContent, m.editCursor)

	case "home":
		m.editCursor = LineStart(m.editContent, m.editCursor)

	case "end":
		m.editCursor = LineEnd(m.editContent, m.editCursor)

	case "ctrl+a":
		// Go to document start
		m.editCursor = 0

	case "ctrl+e":
		// Go to document end
		m.editCursor = len(m.editContent)

	case "ctrl+left":
		// Move word backward
		m.editCursor = WordBackward(m.editContent, m.editCursor)

	case "ctrl+right":
		// Move word forward
		m.editCursor = WordForward(m.editContent, m.editCursor)

	case "ctrl+backspace":
		// Delete word backward
		if m.editCursor > 0 {
			newPos := WordBackward(m.editContent, m.editCursor)
			m.editContent = m.editContent[:newPos] + m.editContent[m.editCursor:]
			m.editCursor = newPos
		}

	case "ctrl+delete":
		// Delete word forward
		if m.editCursor < len(m.editContent) {
			endPos := WordForward(m.editContent, m.editCursor)
			m.editContent = m.editContent[:m.editCursor] + m.editContent[endPos:]
		}

	case "ctrl+v":
		// Paste from clipboard
		if text, err := m.clipboard.Paste(); err == nil && text != "" {
			m.editContent = m.editContent[:m.editCursor] + text + m.editContent[m.editCursor:]
			m.editCursor += len(text)
		}

	case "ctrl+z":
		// Undo
		if content, cursor, ok := m.editHistory.Undo(); ok {
			m.editContent = content
			m.editCursor = cursor
		}

	case "ctrl+y", "ctrl+shift+z":
		// Redo
		if content, cursor, ok := m.editHistory.Redo(); ok {
			m.editContent = content
			m.editCursor = cursor
		}

	case "tab":
		// Insert tab as spaces
		spaces := "    "
		m.editContent = m.editContent[:m.editCursor] + spaces + m.editContent[m.editCursor:]
		m.editCursor += len(spaces)

	default:
		// Handle regular character input (printable ASCII and extended)
		if len(msg.String()) == 1 && msg.String()[0] >= 32 {
			m.editContent = m.editContent[:m.editCursor] + msg.String() + m.editContent[m.editCursor:]
			m.editCursor++
		}
	}

	return m, nil
}

func (m Model) handleVimEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Always handle ctrl+s for save
	if key == "ctrl+s" {
		if m.editContent != "" && m.editingMemo != nil {
			m.loading = true
			return m, m.updateMemo(m.editingMemo.Name, m.editContent)
		}
		return m, nil
	}

	switch m.editorMode {
	case ModeNormal:
		return m.handleVimNormalMode(msg, false)
	case ModeInsert:
		return m.handleVimInsertMode(msg, false)
	case ModeVisual:
		return m.handleVimVisualMode(msg, false)
	}

	return m, nil
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
			m.createContent = ""
			m.createCursor = 0
		} else if m.currentScreen == ScreenEdit {
			m.currentScreen = m.previousScreen
			m.editContent = ""
			m.editOriginal = ""
			m.editCursor = 0
			m.editingMemo = nil
		}

	case "n", "N", "esc":
		// Keep editing
		m.showingUnsavedDialog = false
	}

	return m, nil
}
