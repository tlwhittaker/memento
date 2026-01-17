package ui

// HelpItem represents a single help entry.
type HelpItem struct {
	Key         string
	Description string
}

// HelpSection represents a group of help items.
type HelpSection struct {
	Title string
	Items []HelpItem
}

// GetHelpSections returns context-aware help based on the current screen.
func GetHelpSections(screen Screen, inSplitPane bool, inCalendarFocus bool, editorMode string) []HelpSection {
	switch screen {
	case ScreenList:
		return getListHelp(inSplitPane, inCalendarFocus)
	case ScreenDetail:
		return getDetailHelp()
	case ScreenCreate, ScreenEdit:
		return getEditorHelp(editorMode)
	case ScreenCalendar:
		return getCalendarHelp()
	case ScreenTags:
		return getTagsHelp()
	default:
		return getListHelp(inSplitPane, false)
	}
}

func getListHelp(inSplitPane bool, inCalendarFocus bool) []HelpSection {
	sections := []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{"j/k", "Move down/up"},
				{"g g", "Jump to first memo"},
				{"G", "Jump to last memo"},
				{"Enter", "Open memo detail"},
				{"/", "Search memos"},
				{"Esc", "Clear search"},
			},
		},
		{
			Title: "Actions",
			Items: []HelpItem{
				{"n", "Create new memo"},
				{"e", "Edit current memo"},
				{"d", "Delete current memo"},
				{"p", "Toggle pin"},
				{"a", "Toggle archive"},
				{"v", "Cycle visibility"},
				{"r", "Refresh memos"},
			},
		},
		{
			Title: "Views",
			Items: []HelpItem{
				{"c", "Calendar view"},
				{":", "Command palette"},
				{"?", "Toggle help"},
				{"q", "Quit"},
			},
		},
	}

	if inSplitPane {
		sections = append([]HelpSection{{
			Title: "Split Pane",
			Items: []HelpItem{
				{"h/l", "Switch pane focus"},
				{"Tab", "Next pane"},
			},
		}}, sections...)
	}

	return sections
}

func getDetailHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{"j/k", "Scroll down/up"},
				{"Esc/q", "Back to list"},
			},
		},
		{
			Title: "Actions",
			Items: []HelpItem{
				{"e", "Edit memo"},
				{"d", "Delete memo"},
				{"p", "Toggle pin"},
				{"a", "Toggle archive"},
				{"v", "Cycle visibility"},
				{"m", "Toggle markdown render"},
				{"o", "Open URL under cursor"},
				{"y", "Copy content to clipboard"},
			},
		},
		{
			Title: "General",
			Items: []HelpItem{
				{":", "Command palette"},
				{"?", "Toggle help"},
			},
		},
	}
}

func getEditorHelp(editorMode string) []HelpSection {
	if editorMode == "simple" {
		return []HelpSection{
			{
				Title: "Simple Editor",
				Items: []HelpItem{
					{"Ctrl+s", "Save memo"},
					{"Esc", "Cancel (confirm if unsaved)"},
				},
			},
			{
				Title: "Editing",
				Items: []HelpItem{
					{"Type", "Insert text at cursor"},
					{"Backspace", "Delete character"},
					{"Enter", "New line"},
				},
			},
			{
				Title: "Navigation",
				Items: []HelpItem{
					{"←/→/↑/↓", "Move cursor"},
					{"Home/End", "Line start/end"},
				},
			},
			{
				Title: "Tip",
				Items: []HelpItem{
					{":editor vim", "Switch to vim mode"},
				},
			},
		}
	}

	// Vim mode help
	return []HelpSection{
		{
			Title: "Vim Modes",
			Items: []HelpItem{
				{"i", "Enter insert mode"},
				{"Esc", "Return to normal mode"},
				{"v", "Enter visual mode"},
			},
		},
		{
			Title: "Editing",
			Items: []HelpItem{
				{"dd", "Delete line"},
				{"yy", "Yank line"},
				{"p", "Paste"},
				{"u", "Undo"},
				{"Ctrl+r", "Redo"},
			},
		},
		{
			Title: "Commands",
			Items: []HelpItem{
				{":w", "Save memo"},
				{":q", "Quit (confirm if unsaved)"},
				{":wq", "Save and quit"},
				{"Ctrl+s", "Save memo"},
			},
		},
		{
			Title: "Navigation",
			Items: []HelpItem{
				{"h/j/k/l", "Move cursor"},
				{"w/b", "Word forward/back"},
				{"0/$", "Line start/end"},
				{"gg/G", "File start/end"},
			},
		},
	}
}

func getCalendarHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{"h/l", "Previous/next day"},
				{"j/k", "Previous/next week"},
				{"H/L", "Previous/next month"},
				{"Enter", "View memos for day"},
				{"Esc/q", "Back to list"},
			},
		},
		{
			Title: "Actions",
			Items: []HelpItem{
				{"n", "Create new memo"},
				{"w", "Toggle week view"},
				{":", "Command palette"},
				{"?", "Toggle help"},
			},
		},
	}
}

func getTagsHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{"j/k", "Move down/up"},
				{"Enter", "Filter by tag"},
				{"Esc/q", "Back to list"},
			},
		},
		{
			Title: "General",
			Items: []HelpItem{
				{":", "Command palette"},
				{"?", "Toggle help"},
			},
		},
	}
}
