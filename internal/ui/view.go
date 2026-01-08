package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current state of the model.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content string

	switch m.currentScreen {
	case ScreenList:
		content = m.renderListScreen()
	case ScreenDetail:
		content = m.renderDetailScreen()
	case ScreenCreate:
		content = m.renderCreateScreen()
	case ScreenEdit:
		content = m.renderEditScreen()
	}

	// Overlay dialogs
	if m.confirmingDelete {
		content = m.overlayDeleteConfirmation(content)
	}
	if m.showingUnsavedDialog {
		content = m.overlayUnsavedConfirmation(content)
	}

	return content
}

func (m Model) renderListScreen() string {
	// Calculate dimensions
	contentWidth := m.width - 2
	contentHeight := m.height - 4

	var contentBuilder strings.Builder

	// Header
	header := HeaderStyle.Render("Memos")
	contentBuilder.WriteString(header)
	contentBuilder.WriteString("\n")

	// Search bar (if active or has query)
	if m.searchActive || m.searchQuery != "" {
		searchPrefix := MutedStyle.Render("/")
		searchContent := m.searchQuery
		if m.searchActive {
			// Show cursor in search
			if m.searchCursor >= len(searchContent) {
				searchContent = searchContent + CursorStyle.Render("_")
			} else {
				searchContent = searchContent[:m.searchCursor] +
					CursorStyle.Render(string(searchContent[m.searchCursor])) +
					searchContent[m.searchCursor+1:]
			}
		}
		searchLine := searchPrefix + searchContent
		if m.searchQuery != "" && m.filteredMemos != nil {
			searchLine += MutedStyle.Render(fmt.Sprintf(" (%d matches)", len(m.filteredMemos)))
		}
		contentBuilder.WriteString(searchLine)
		contentBuilder.WriteString("\n")
	}

	// Error or status message
	if m.err != nil {
		contentBuilder.WriteString(ErrorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())))
		contentBuilder.WriteString("\n")
	} else if m.statusMessage != "" {
		contentBuilder.WriteString(SuccessStyle.Render(m.statusMessage))
		contentBuilder.WriteString("\n")
	}

	// Loading indicator
	if m.loading {
		contentBuilder.WriteString(MutedStyle.Render("Loading..."))
		contentBuilder.WriteString("\n")
	}

	// Memo list
	listHeight := contentHeight - 5
	if m.err != nil || m.statusMessage != "" {
		listHeight--
	}
	if m.searchActive || m.searchQuery != "" {
		listHeight--
	}

	// Get display memos (filtered or all)
	displayMemos := m.memos
	if m.searchQuery != "" && m.filteredMemos != nil {
		displayMemos = m.filteredMemos
	}

	if len(displayMemos) == 0 && !m.loading {
		contentBuilder.WriteString("\n")
		if m.searchQuery != "" {
			contentBuilder.WriteString(MutedStyle.Render("  No matching memos found."))
		} else {
			contentBuilder.WriteString(MutedStyle.Render("  No memos found. Press 'n' to create one."))
		}
		contentBuilder.WriteString("\n")
	} else {
		offset := m.listOffset
		if m.listCursor < offset {
			offset = m.listCursor
		}
		if m.listCursor >= offset+listHeight {
			offset = m.listCursor - listHeight + 1
		}

		for i := 0; i < listHeight && i+offset < len(displayMemos); i++ {
			idx := i + offset
			memo := displayMemos[idx]
			isSelected := idx == m.listCursor

			// Build line - single preview format
			var lineBuilder strings.Builder

			// Selection indicator
			if isSelected {
				lineBuilder.WriteString(MemoSelectedStyle.Render(" > "))
			} else {
				lineBuilder.WriteString("   ")
			}

			// Pinned indicator at start
			if memo.Pinned {
				lineBuilder.WriteString(PinnedStyle.Render("* "))
			}

			// Preview (truncated content on single line)
			date := memo.ShortDate()
			maxPreviewLen := contentWidth - 20 - len(date)
			if memo.Pinned {
				maxPreviewLen -= 2
			}
			preview := memo.Preview(maxPreviewLen)

			if isSelected {
				lineBuilder.WriteString(MemoSelectedBgStyle.Render(preview))
			} else {
				lineBuilder.WriteString(MemoPreviewStyle.Render(preview))
			}

			// Date (right aligned)
			currentLen := lipgloss.Width(lineBuilder.String())
			padding := contentWidth - currentLen - len(date) - 2
			if padding > 0 {
				lineBuilder.WriteString(strings.Repeat(" ", padding))
			}
			lineBuilder.WriteString(MemoDateStyle.Render(date))

			contentBuilder.WriteString(lineBuilder.String())
			contentBuilder.WriteString("\n")
		}
	}

	// Build final content with box
	content := contentBuilder.String()

	// Status bar
	var leftStatus, rightHelp string
	if m.searchActive {
		leftStatus = fmt.Sprintf("%d/%d memos", len(displayMemos), len(m.memos))
		rightHelp = "enter:confirm  esc:cancel  /:search"
	} else {
		leftStatus = fmt.Sprintf("%d memos", len(displayMemos))
		rightHelp = "n:new  e:edit  d:del  /:search  q:quit"
	}
	statusBar := m.renderStatusBar(leftStatus, rightHelp)

	// Combine with border
	lines := strings.Split(content, "\n")
	for len(lines) < contentHeight-1 {
		lines = append(lines, "")
	}
	content = strings.Join(lines[:contentHeight-1], "\n")

	box := BoxStyle.Width(contentWidth).Render(content)
	return box + "\n" + statusBar
}

func (m Model) renderDetailScreen() string {
	if m.selectedMemo == nil {
		return "No memo selected"
	}

	contentWidth := m.width - 2
	contentHeight := m.height - 4

	var contentBuilder strings.Builder

	// Header with memo info
	header := fmt.Sprintf("Memo #%s", m.selectedMemo.ShortID())
	contentBuilder.WriteString(HeaderStyle.Render(header))
	contentBuilder.WriteString("\n")

	// Metadata line
	meta := MutedStyle.Render(fmt.Sprintf("Created: %s", m.selectedMemo.FormattedDate()))
	contentBuilder.WriteString(meta)
	contentBuilder.WriteString("\n")
	contentBuilder.WriteString(MutedStyle.Render(strings.Repeat("─", contentWidth-4)))
	contentBuilder.WriteString("\n")

	// Error message
	if m.err != nil {
		contentBuilder.WriteString(ErrorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())))
		contentBuilder.WriteString("\n")
	}

	// Content with scrolling - wrap long lines at word boundaries
	rawLines := strings.Split(m.selectedMemo.Content, "\n")
	wrapWidth := contentWidth - 6
	var contentLines []string
	for _, line := range rawLines {
		for len(line) > wrapWidth {
			breakPoint := strings.LastIndex(line[:wrapWidth], " ")
			if breakPoint <= 0 {
				breakPoint = wrapWidth
			}
			contentLines = append(contentLines, line[:breakPoint])
			line = strings.TrimLeft(line[breakPoint:], " ")
		}
		contentLines = append(contentLines, line)
	}
	availableHeight := contentHeight - 7

	maxScroll := len(contentLines) - availableHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.detailScroll > maxScroll {
		m.detailScroll = maxScroll
	}
	if m.detailScroll < 0 {
		m.detailScroll = 0
	}

	endIdx := m.detailScroll + availableHeight
	if endIdx > len(contentLines) {
		endIdx = len(contentLines)
	}

	for i := m.detailScroll; i < endIdx; i++ {
		contentBuilder.WriteString(contentLines[i])
		contentBuilder.WriteString("\n")
	}

	// Build content
	content := contentBuilder.String()
	lines := strings.Split(content, "\n")
	for len(lines) < contentHeight-1 {
		lines = append(lines, "")
	}
	content = strings.Join(lines[:contentHeight-1], "\n")

	// Status bar with scroll info
	scrollInfo := ""
	if len(contentLines) > availableHeight {
		scrollInfo = fmt.Sprintf("Line %d/%d", m.detailScroll+1, len(contentLines))
	}
	statusBar := m.renderStatusBar(scrollInfo, "e:edit  d:delete  esc:back")

	box := BoxStyle.Width(contentWidth).Render(content)
	return box + "\n" + statusBar
}

func (m Model) renderCreateScreen() string {
	return m.renderEditorScreen("Create New Memo", m.createContent, m.createCursor, false)
}

func (m Model) renderEditScreen() string {
	title := "Edit Memo"
	if m.editingMemo != nil {
		title = fmt.Sprintf("Edit Memo #%s", m.editingMemo.ShortID())
	}
	return m.renderEditorScreen(title, m.editContent, m.editCursor, true)
}

func (m Model) renderEditorScreen(title, content string, cursor int, isEdit bool) string {
	contentWidth := m.width - 2
	contentHeight := m.height - 4
	textAreaHeight := contentHeight - 5

	var mainBuilder strings.Builder

	// Header
	mainBuilder.WriteString(HeaderStyle.Render(title))
	mainBuilder.WriteString("\n\n")

	// Calculate cursor position
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}

	cursorLine := 0
	cursorCol := cursor
	pos := 0
	for i, line := range lines {
		lineEnd := pos + len(line)
		if cursor <= lineEnd {
			cursorLine = i
			cursorCol = cursor - pos
			break
		}
		pos = lineEnd + 1 // +1 for newline
		cursorLine = i + 1
		cursorCol = 0
	}

	// Render text area content
	textAreaWidth := contentWidth - 6
	var textContent strings.Builder

	for i := 0; i < textAreaHeight; i++ {
		lineContent := ""
		if i < len(lines) {
			lineContent = lines[i]
		}

		// Truncate if needed
		displayLine := lineContent
		if len(displayLine) > textAreaWidth {
			displayLine = displayLine[:textAreaWidth]
		}

		// Insert cursor
		if i == cursorLine && !m.loading {
			if cursorCol >= len(displayLine) {
				displayLine = displayLine + CursorStyle.Render("_")
			} else if cursorCol < len(displayLine) {
				displayLine = displayLine[:cursorCol] + CursorStyle.Render(string(displayLine[cursorCol])) + displayLine[cursorCol+1:]
			}
		}

		textContent.WriteString(displayLine)
		if i < textAreaHeight-1 {
			textContent.WriteString("\n")
		}
	}

	// Create text area box
	textAreaBox := TextAreaStyle.Width(textAreaWidth + 2).Height(textAreaHeight).Render(textContent.String())
	mainBuilder.WriteString(textAreaBox)

	// Build main content
	mainContent := mainBuilder.String()
	mainLines := strings.Split(mainContent, "\n")
	for len(mainLines) < contentHeight-1 {
		mainLines = append(mainLines, "")
	}
	mainContent = strings.Join(mainLines[:contentHeight-1], "\n")

	// Status bar with position info and vim mode
	charCount := fmt.Sprintf("%d chars", len(content))
	posInfo := fmt.Sprintf("Ln %d, Col %d", cursorLine+1, cursorCol+1)
	leftStatus := charCount + StatusSeparator.String() + posInfo

	// Add vim mode indicator if vim mode is enabled
	rightHelp := "ctrl+s:save  esc:cancel"
	if m.settings.IsVimMode() {
		var modeIndicator string
		switch m.editorMode {
		case ModeNormal:
			modeIndicator = NormalModeStyle.Render("NORMAL")
		case ModeInsert:
			modeIndicator = InsertModeStyle.Render("INSERT")
		case ModeVisual:
			modeIndicator = VisualModeStyle.Render("VISUAL")
		}
		leftStatus = modeIndicator + " " + leftStatus
		rightHelp = "ctrl+s:save  i:insert  esc:normal"
	}

	statusBar := m.renderStatusBar(leftStatus, rightHelp)

	box := BoxStyle.Width(contentWidth).Render(mainContent)
	return box + "\n" + statusBar
}

func (m Model) renderStatusBar(left, right string) string {
	leftStyled := StatusBarStyle.Render(left)
	rightStyled := HelpStyle.Render(right)

	gap := m.width - lipgloss.Width(leftStyled) - lipgloss.Width(rightStyled) - 2
	if gap < 0 {
		gap = 0
	}

	return " " + leftStyled + strings.Repeat(" ", gap) + rightStyled
}

func (m Model) overlayDeleteConfirmation(content string) string {
	memoTitle := ""
	if m.selectedMemo != nil {
		memoTitle = m.selectedMemo.Title()
		if len(memoTitle) > 25 {
			memoTitle = memoTitle[:22] + "..."
		}
	}

	dialogContent := DialogTitleStyle.Render("Delete Memo?") + "\n\n" +
		fmt.Sprintf("Delete \"%s\"?\n\n", memoTitle) +
		StatusKeyStyle.Render("y") + " confirm  " +
		StatusKeyStyle.Render("n") + " cancel"

	dialog := DialogStyle.Width(36).Render(dialogContent)
	return m.centerOverlay(content, dialog)
}

func (m Model) overlayUnsavedConfirmation(content string) string {
	dialogContent := UnsavedDialogTitleStyle.Render("Unsaved Changes") + "\n\n" +
		"Discard your changes?\n\n" +
		StatusKeyStyle.Render("y") + " discard  " +
		StatusKeyStyle.Render("n") + " keep editing"

	dialog := UnsavedDialogStyle.Width(36).Render(dialogContent)
	return m.centerOverlay(content, dialog)
}

func (m Model) centerOverlay(background, overlay string) string {
	bgLines := strings.Split(background, "\n")
	overlayLines := strings.Split(overlay, "\n")

	startY := (m.height - len(overlayLines)) / 2
	startX := (m.width - lipgloss.Width(overlayLines[0])) / 2

	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	// Overlay dialog on background
	for i, overlayLine := range overlayLines {
		bgIdx := startY + i
		if bgIdx >= 0 && bgIdx < len(bgLines) {
			bgLine := bgLines[bgIdx]

			// Ensure background line is long enough
			for len(bgLine) < startX+lipgloss.Width(overlayLine) {
				bgLine += " "
			}

			// Splice in overlay line
			newLine := ""
			if startX > 0 && len(bgLine) >= startX {
				newLine = bgLine[:startX]
			} else {
				newLine = strings.Repeat(" ", startX)
			}
			newLine += overlayLine

			overlayWidth := lipgloss.Width(overlayLine)
			if startX+overlayWidth < len(bgLine) {
				newLine += bgLine[startX+overlayWidth:]
			}

			bgLines[bgIdx] = newLine
		}
	}

	return strings.Join(bgLines, "\n")
}
