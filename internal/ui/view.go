package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tlwhittaker/memento/internal/models"
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
	case ScreenCalendar:
		content = m.renderCalendarScreen()
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
	if m.width >= SplitPaneMinWidth {
		return m.renderSplitPaneList()
	}
	return m.renderSinglePaneList()
}

func (m Model) renderSplitPaneList() string {
	leftWidth := (m.width - 4) / 2
	rightWidth := m.width - leftWidth - 4
	contentHeight := m.height - 4

	leftPane := m.renderMemoListPane(leftWidth, contentHeight, !m.splitFocusRight)
	rightPane := m.renderPreviewPane(rightWidth, contentHeight, m.splitFocusRight)

	combined := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	displayMemos := m.getDisplayMemos()
	showInlineCal := m.shouldShowInlineCalendar()
	var leftStatus, rightHelp string
	if m.searchActive {
		leftStatus = fmt.Sprintf("%d/%d memos", len(displayMemos), len(m.memos))
		rightHelp = "enter:open  /:search  q:quit"
	} else if m.splitFocusRight {
		leftStatus = fmt.Sprintf("%d memos", len(displayMemos))
		rightHelp = "←:list  j/k:scroll  e:edit  q:quit"
	} else if showInlineCal && m.inlineCalendarFocus {
		leftStatus = fmt.Sprintf("%d memos", len(displayMemos))
		rightHelp = "h/l:day  j/k:week  H/L:month  c:list  →:preview"
	} else {
		leftStatus = fmt.Sprintf("%d memos", len(displayMemos))
		if showInlineCal {
			rightHelp = "→:preview  c:calendar  n:new  e:edit  /:search"
		} else {
			rightHelp = "→:preview  n:new  e:edit  /:search  c:cal  q:quit"
		}
	}
	statusBar := m.renderStatusBar(leftStatus, rightHelp)

	return combined + "\n" + statusBar
}

const (
	InlineCalendarMinHeight = 32 // Minimum height to show inline calendar
	InlineCalendarBoxHeight = 12 // Height of the calendar box (including border)
)

func (m Model) shouldShowInlineCalendar() bool {
	return m.width >= SplitPaneMinWidth && m.height >= InlineCalendarMinHeight
}

func (m Model) renderMemoListPane(width, height int, focused bool) string {
	showInlineCal := m.shouldShowInlineCalendar()
	listFocused := focused && !m.inlineCalendarFocus
	calFocused := focused && m.inlineCalendarFocus
	tz := m.settings.GetTimezone()

	// Calculate heights
	calendarHeight := 0
	if showInlineCal {
		calendarHeight = InlineCalendarBoxHeight
	}
	listBoxHeight := height - calendarHeight

	// Build list content
	var listBuilder strings.Builder

	header := HeaderStyle.Render("Memos")
	listBuilder.WriteString(header)
	listBuilder.WriteString("\n")

	if m.searchActive || m.searchQuery != "" {
		searchPrefix := MutedStyle.Render("/")
		searchContent := m.searchQuery
		if m.searchActive {
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
			searchLine += MutedStyle.Render(fmt.Sprintf(" (%d)", len(m.filteredMemos)))
		}
		listBuilder.WriteString(searchLine)
		listBuilder.WriteString("\n")
	}

	if m.err != nil {
		listBuilder.WriteString(ErrorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())))
		listBuilder.WriteString("\n")
	} else if m.statusMessage != "" {
		listBuilder.WriteString(SuccessStyle.Render(m.statusMessage))
		listBuilder.WriteString("\n")
	}

	// Calculate available lines for memo list
	listContentHeight := listBoxHeight - 4 // Account for box borders and header
	if m.err != nil || m.statusMessage != "" {
		listContentHeight--
	}
	if m.searchActive || m.searchQuery != "" {
		listContentHeight--
	}

	displayMemos := m.getDisplayMemos()

	if len(displayMemos) == 0 && !m.loading {
		listBuilder.WriteString("\n")
		if m.searchQuery != "" {
			listBuilder.WriteString(MutedStyle.Render("  No matches"))
		} else {
			listBuilder.WriteString(MutedStyle.Render("  No memos. Press 'n'"))
		}
		listBuilder.WriteString("\n")
	} else {
		offset := m.listOffset
		if m.listCursor < offset {
			offset = m.listCursor
		}
		if m.listCursor >= offset+listContentHeight {
			offset = m.listCursor - listContentHeight + 1
		}

		for i := 0; i < listContentHeight && i+offset < len(displayMemos); i++ {
			idx := i + offset
			memo := displayMemos[idx]
			isSelected := idx == m.listCursor

			var lineBuilder strings.Builder
			if isSelected && listFocused {
				lineBuilder.WriteString(MemoSelectedStyle.Render(" > "))
			} else if isSelected {
				lineBuilder.WriteString(MutedStyle.Render(" > "))
			} else {
				lineBuilder.WriteString("   ")
			}

			if memo.Pinned {
				lineBuilder.WriteString(PinnedStyle.Render("* "))
			}

			date := memo.ShortDateIn(tz)
			maxPreviewLen := width - 16 - len(date)
			if memo.Pinned {
				maxPreviewLen -= 2
			}
			if maxPreviewLen < 10 {
				maxPreviewLen = 10
			}
			preview := memo.Preview(maxPreviewLen)

			if isSelected && listFocused {
				lineBuilder.WriteString(MemoSelectedBgStyle.Render(preview))
			} else if isSelected {
				lineBuilder.WriteString(MutedStyle.Render(preview))
			} else {
				lineBuilder.WriteString(MemoPreviewStyle.Render(preview))
			}

			currentLen := lipgloss.Width(lineBuilder.String())
			padding := width - currentLen - len(date) - 4
			if padding > 0 {
				lineBuilder.WriteString(strings.Repeat(" ", padding))
			}
			lineBuilder.WriteString(MemoDateStyle.Render(date))

			listBuilder.WriteString(lineBuilder.String())
			listBuilder.WriteString("\n")
		}
	}

	// Render list box
	listContent := listBuilder.String()
	listLines := strings.Split(listContent, "\n")
	for len(listLines) < listBoxHeight-3 {
		listLines = append(listLines, "")
	}
	listContent = strings.Join(listLines[:listBoxHeight-3], "\n")

	listBoxStyle := ContentBoxStyle
	if listFocused {
		listBoxStyle = BoxStyle
	}
	listBox := listBoxStyle.Width(width).Render(listContent)

	if !showInlineCal {
		return listBox
	}

	// Render calendar box
	calendarBox := m.renderInlineCalendarBox(width, calendarHeight, calFocused)

	return listBox + "\n" + calendarBox
}

func (m Model) renderInlineCalendarBox(width, height int, focused bool) string {
	tz := m.settings.GetTimezone()
	var b strings.Builder

	currentDate := time.Date(m.calendarYear, time.Month(m.calendarMonth), 1, 0, 0, 0, 0, tz)
	monthName := currentDate.Format("Jan 2006")

	// Day headers
	b.WriteString(MutedStyle.Render(" Su Mo Tu We Th Fr Sa"))
	b.WriteString("\n")

	memoDates := m.getMemoDates()
	firstDay := time.Date(m.calendarYear, time.Month(m.calendarMonth), 1, 0, 0, 0, 0, tz)
	startWeekday := int(firstDay.Weekday())
	daysInMonth := time.Date(m.calendarYear, time.Month(m.calendarMonth)+1, 0, 0, 0, 0, 0, tz).Day()

	today := time.Now().In(tz)
	isCurrentMonth := today.Year() == m.calendarYear && int(today.Month()) == m.calendarMonth

	// Leading spaces
	for i := 0; i < startWeekday; i++ {
		b.WriteString("   ")
	}

	for day := 1; day <= daysInMonth; day++ {
		dateKey := fmt.Sprintf("%04d-%02d-%02d", m.calendarYear, m.calendarMonth, day)
		hasMemo := memoDates[dateKey]
		isSelected := day == m.calendarDay && focused
		isToday := isCurrentMonth && day == today.Day()

		dayStr := fmt.Sprintf("%2d", day)

		if hasMemo {
			dayStr += "·"
		} else {
			dayStr += " "
		}

		if isSelected {
			b.WriteString(MemoSelectedBgStyle.Render(dayStr))
		} else if isToday {
			b.WriteString(lipgloss.NewStyle().Foreground(SecondaryColor).Bold(true).Render(dayStr))
		} else if hasMemo {
			b.WriteString(lipgloss.NewStyle().Foreground(PrimaryColor).Render(dayStr))
		} else {
			b.WriteString(MutedStyle.Render(dayStr))
		}

		if (startWeekday+day)%7 == 0 {
			b.WriteString("\n")
		}
	}

	// Selected date info
	b.WriteString("\n")
	selectedDate := time.Date(m.calendarYear, time.Month(m.calendarMonth), m.calendarDay, 0, 0, 0, 0, tz)
	dayMemos := m.getMemosForDate(m.calendarYear, m.calendarMonth, m.calendarDay)
	dateInfo := selectedDate.Format("Mon Jan 2")
	if len(dayMemos) > 0 {
		dateInfo += fmt.Sprintf(" (%d memo", len(dayMemos))
		if len(dayMemos) > 1 {
			dateInfo += "s"
		}
		dateInfo += ")"
	}
	if focused {
		b.WriteString(SubtitleStyle.Render(dateInfo))
	} else {
		b.WriteString(MutedStyle.Render(dateInfo))
	}

	calContent := b.String()

	// Center the calendar content
	calLines := strings.Split(calContent, "\n")
	boxInnerWidth := width - 4

	var centeredLines []string
	for _, line := range calLines {
		lineWidth := lipgloss.Width(line)
		leftPad := (boxInnerWidth - lineWidth) / 2
		if leftPad < 0 {
			leftPad = 0
		}
		centeredLines = append(centeredLines, strings.Repeat(" ", leftPad)+line)
	}

	// Pad to fill height
	contentHeight := height - 3
	for len(centeredLines) < contentHeight {
		centeredLines = append(centeredLines, "")
	}
	centeredContent := strings.Join(centeredLines[:contentHeight], "\n")

	// Create box with month header
	boxStyle := ContentBoxStyle
	if focused {
		boxStyle = BoxStyle
	}

	// Add month name as header
	headerStyle := MutedStyle
	if focused {
		headerStyle = SubtitleStyle
	}
	header := headerStyle.Render("─ " + monthName + " ─")

	// Calculate centering for header
	headerWidth := lipgloss.Width(header)
	headerPad := (boxInnerWidth - headerWidth) / 2
	if headerPad < 0 {
		headerPad = 0
	}
	centeredHeader := strings.Repeat(" ", headerPad) + header

	return boxStyle.Width(width).Render(centeredHeader + "\n" + centeredContent)
}

func (m Model) renderPreviewPane(width, height int, focused bool) string {
	var contentBuilder strings.Builder
	tz := m.settings.GetTimezone()

	displayMemos := m.getDisplayMemos()
	if len(displayMemos) == 0 || m.listCursor >= len(displayMemos) {
		contentBuilder.WriteString(HeaderStyle.Render("Preview"))
		contentBuilder.WriteString("\n\n")
		contentBuilder.WriteString(MutedStyle.Render("  No memo selected"))
	} else {
		memo := displayMemos[m.listCursor]

		header := fmt.Sprintf("Memo #%s", memo.ShortID())
		contentBuilder.WriteString(HeaderStyle.Render(header))
		contentBuilder.WriteString("\n")

		meta := MutedStyle.Render(memo.FormattedDateIn(tz))
		contentBuilder.WriteString(meta)
		contentBuilder.WriteString("\n")

		if memo.Pinned {
			contentBuilder.WriteString(PinnedStyle.Render("Pinned "))
		}
		contentBuilder.WriteString(MutedStyle.Render(memo.Visibility))
		contentBuilder.WriteString("\n")
		contentBuilder.WriteString(MutedStyle.Render(strings.Repeat("─", width-4)))
		contentBuilder.WriteString("\n")

		wrapWidth := width - 6
		rawLines := strings.Split(memo.Content, "\n")
		var wrappedLines []string
		for _, line := range rawLines {
			for len(line) > wrapWidth {
				breakPoint := strings.LastIndex(line[:wrapWidth], " ")
				if breakPoint <= 0 {
					breakPoint = wrapWidth
				}
				wrappedLines = append(wrappedLines, line[:breakPoint])
				line = strings.TrimLeft(line[breakPoint:], " ")
			}
			wrappedLines = append(wrappedLines, line)
		}

		availableHeight := height - 7

		startLine := m.previewScroll
		if startLine > len(wrappedLines)-availableHeight {
			startLine = len(wrappedLines) - availableHeight
		}
		if startLine < 0 {
			startLine = 0
		}

		endLine := startLine + availableHeight
		if endLine > len(wrappedLines) {
			endLine = len(wrappedLines)
		}

		for i := startLine; i < endLine; i++ {
			contentBuilder.WriteString(wrappedLines[i])
			contentBuilder.WriteString("\n")
		}

		if len(wrappedLines) > availableHeight {
			scrollInfo := fmt.Sprintf("Line %d/%d", startLine+1, len(wrappedLines))
			contentBuilder.WriteString(MutedStyle.Render(scrollInfo))
		}
	}

	content := contentBuilder.String()
	lines := strings.Split(content, "\n")
	for len(lines) < height-1 {
		lines = append(lines, "")
	}
	content = strings.Join(lines[:height-1], "\n")

	boxStyle := ContentBoxStyle
	if focused {
		boxStyle = BoxStyle
	}
	return boxStyle.Width(width).Render(content)
}

func (m Model) renderSinglePaneList() string {
	contentWidth := m.width - 2
	contentHeight := m.height - 4
	tz := m.settings.GetTimezone()

	var contentBuilder strings.Builder

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
			date := memo.ShortDateIn(tz)
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

	var leftStatus, rightHelp string
	if m.searchActive {
		leftStatus = fmt.Sprintf("%d/%d memos", len(displayMemos), len(m.memos))
		rightHelp = "enter:confirm  esc:cancel  /:search"
	} else {
		leftStatus = fmt.Sprintf("%d memos", len(displayMemos))
		rightHelp = "n:new  e:edit  d:del  /:search  c:cal  q:quit"
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
	tz := m.settings.GetTimezone()

	var contentBuilder strings.Builder

	// Header with memo info
	header := fmt.Sprintf("Memo #%s", m.selectedMemo.ShortID())
	contentBuilder.WriteString(HeaderStyle.Render(header))
	contentBuilder.WriteString("\n")

	// Metadata line
	meta := MutedStyle.Render(fmt.Sprintf("Created: %s", m.selectedMemo.FormattedDateIn(tz)))
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

func (m Model) renderCalendarScreen() string {
	contentWidth := m.width - 2
	contentHeight := m.height - 4
	tz := m.settings.GetTimezone()

	var contentBuilder strings.Builder

	currentDate := time.Date(m.calendarYear, time.Month(m.calendarMonth), 1, 0, 0, 0, 0, tz)
	monthName := currentDate.Format("January 2006")

	header := HeaderStyle.Render("Calendar: " + monthName)
	contentBuilder.WriteString(header)
	contentBuilder.WriteString("\n\n")

	dayHeaders := " Su  Mo  Tu  We  Th  Fr  Sa"
	contentBuilder.WriteString(MutedStyle.Render(dayHeaders))
	contentBuilder.WriteString("\n")

	memoDates := m.getMemoDates()

	firstDay := time.Date(m.calendarYear, time.Month(m.calendarMonth), 1, 0, 0, 0, 0, tz)
	startWeekday := int(firstDay.Weekday())

	daysInMonth := time.Date(m.calendarYear, time.Month(m.calendarMonth)+1, 0, 0, 0, 0, 0, tz).Day()

	today := time.Now().In(tz)
	isCurrentMonth := today.Year() == m.calendarYear && int(today.Month()) == m.calendarMonth

	for i := 0; i < startWeekday; i++ {
		contentBuilder.WriteString("    ")
	}

	for day := 1; day <= daysInMonth; day++ {
		dateKey := fmt.Sprintf("%04d-%02d-%02d", m.calendarYear, m.calendarMonth, day)
		hasMemo := memoDates[dateKey]
		isSelected := day == m.calendarDay
		isToday := isCurrentMonth && day == today.Day()

		dayStr := fmt.Sprintf("%3d", day)

		if hasMemo {
			dayStr += "·"
		} else {
			dayStr += " "
		}

		if isSelected {
			contentBuilder.WriteString(MemoSelectedBgStyle.Render(dayStr))
		} else if isToday {
			contentBuilder.WriteString(lipgloss.NewStyle().Foreground(SecondaryColor).Bold(true).Render(dayStr))
		} else if hasMemo {
			contentBuilder.WriteString(lipgloss.NewStyle().Foreground(PrimaryColor).Render(dayStr))
		} else {
			contentBuilder.WriteString(dayStr)
		}

		if (startWeekday+day)%7 == 0 {
			contentBuilder.WriteString("\n")
		}
	}
	contentBuilder.WriteString("\n\n")

	contentBuilder.WriteString(MutedStyle.Render(strings.Repeat("─", contentWidth-4)))
	contentBuilder.WriteString("\n")

	selectedDate := time.Date(m.calendarYear, time.Month(m.calendarMonth), m.calendarDay, 0, 0, 0, 0, tz)
	contentBuilder.WriteString(SubtitleStyle.Render(selectedDate.Format("Monday, January 2, 2006")))
	contentBuilder.WriteString("\n\n")

	dayMemos := m.getMemosForDate(m.calendarYear, m.calendarMonth, m.calendarDay)

	if len(dayMemos) == 0 {
		contentBuilder.WriteString(MutedStyle.Render("  No memos for this day"))
		contentBuilder.WriteString("\n")
	} else {
		availableHeight := contentHeight - 14
		for i, memo := range dayMemos {
			if i >= availableHeight {
				contentBuilder.WriteString(MutedStyle.Render(fmt.Sprintf("  ... +%d more memos", len(dayMemos)-i)))
				contentBuilder.WriteString("\n")
				break
			}
			preview := memo.Preview(contentWidth - 10)
			contentBuilder.WriteString("  • ")
			contentBuilder.WriteString(MemoPreviewStyle.Render(preview))
			contentBuilder.WriteString("\n")
		}
	}

	content := contentBuilder.String()
	lines := strings.Split(content, "\n")
	for len(lines) < contentHeight-1 {
		lines = append(lines, "")
	}
	content = strings.Join(lines[:contentHeight-1], "\n")

	leftStatus := fmt.Sprintf("%d memos this month", m.countMemosInMonth(m.calendarYear, m.calendarMonth))
	rightHelp := "h/l:day  j/k:week  H/L:month  enter:view  esc:back"
	statusBar := m.renderStatusBar(leftStatus, rightHelp)

	box := BoxStyle.Width(contentWidth).Render(content)
	return box + "\n" + statusBar
}

func (m Model) getMemoDates() map[string]bool {
	tz := m.settings.GetTimezone()
	dates := make(map[string]bool)
	for _, memo := range m.memos {
		localTime := memo.DisplayTime.In(tz)
		dateKey := localTime.Format("2006-01-02")
		dates[dateKey] = true
	}
	return dates
}

func (m Model) getMemosForDate(year, month, day int) []models.Memo {
	tz := m.settings.GetTimezone()
	var result []models.Memo
	for _, memo := range m.memos {
		localTime := memo.DisplayTime.In(tz)
		if localTime.Year() == year &&
			int(localTime.Month()) == month &&
			localTime.Day() == day {
			result = append(result, memo)
		}
	}
	return result
}

func (m Model) countMemosInMonth(year, month int) int {
	tz := m.settings.GetTimezone()
	count := 0
	for _, memo := range m.memos {
		localTime := memo.DisplayTime.In(tz)
		if localTime.Year() == year && int(localTime.Month()) == month {
			count++
		}
	}
	return count
}
