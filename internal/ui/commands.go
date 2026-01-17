package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sahilm/fuzzy"
)

// Command represents an executable command in the palette.
type Command struct {
	Name        string
	Description string
	Category    string
	Action      func(m *Model) tea.Cmd
}

// commandSource implements fuzzy.Source for commands.
type commandSource struct {
	commands []Command
}

func (s commandSource) String(i int) string {
	return s.commands[i].Name + " " + s.commands[i].Description
}

func (s commandSource) Len() int {
	return len(s.commands)
}

// GetCommands returns all available commands.
func GetCommands() []Command {
	return []Command{
		// Navigation
		{
			Name:        "list",
			Description: "Go to memo list",
			Category:    "Navigation",
			Action: func(m *Model) tea.Cmd {
				m.currentScreen = ScreenList
				return nil
			},
		},
		{
			Name:        "calendar",
			Description: "Go to calendar view",
			Category:    "Navigation",
			Action: func(m *Model) tea.Cmd {
				m.currentScreen = ScreenCalendar
				return nil
			},
		},
		{
			Name:        "tags",
			Description: "Go to tags browser",
			Category:    "Navigation",
			Action: func(m *Model) tea.Cmd {
				m.currentScreen = ScreenTags
				return nil
			},
		},
		// Actions
		{
			Name:        "new",
			Description: "Create a new memo",
			Category:    "Actions",
			Action: func(m *Model) tea.Cmd {
				m.clearSearch()
				if len(m.templates) > 0 {
					m.showingTemplatePicker = true
					m.templateCursor = 0
				} else {
					m.initCreateEditor("")
					m.currentScreen = ScreenCreate
				}
				return nil
			},
		},
		{
			Name:        "delete",
			Description: "Delete current memo",
			Category:    "Actions",
			Action: func(m *Model) tea.Cmd {
				memoList := m.getDisplayMemos()
				if len(memoList) > 0 && m.listCursor < len(memoList) {
					m.selectedMemo = &memoList[m.listCursor]
					m.confirmingDelete = true
				}
				return nil
			},
		},
		{
			Name:        "pin",
			Description: "Toggle pin on current memo",
			Category:    "Actions",
			Action: func(m *Model) tea.Cmd {
				memoList := m.getDisplayMemos()
				if len(memoList) > 0 && m.listCursor < len(memoList) {
					memo := memoList[m.listCursor]
					m.loading = true
					return m.togglePin(memo.Name, memo.Pinned)
				}
				return nil
			},
		},
		{
			Name:        "archive",
			Description: "Toggle archive on current memo",
			Category:    "Actions",
			Action: func(m *Model) tea.Cmd {
				memoList := m.getDisplayMemos()
				if len(memoList) > 0 && m.listCursor < len(memoList) {
					memo := memoList[m.listCursor]
					m.loading = true
					return m.toggleArchive(memo.Name, memo.IsArchived())
				}
				return nil
			},
		},
		{
			Name:        "refresh",
			Description: "Reload memos from server",
			Category:    "Actions",
			Action: func(m *Model) tea.Cmd {
				m.loading = true
				m.clearSearch()
				return m.loadMemos()
			},
		},
		// Filters
		{
			Name:        "pinned",
			Description: "Show pinned memos",
			Category:    "Filters",
			Action: func(m *Model) tea.Cmd {
				m.searchQuery = "pinned"
				m.applyAdvancedFilter()
				return nil
			},
		},
		{
			Name:        "archived",
			Description: "Show archived memos",
			Category:    "Filters",
			Action: func(m *Model) tea.Cmd {
				m.searchQuery = "archived"
				m.applyAdvancedFilter()
				return nil
			},
		},
		{
			Name:        "clear-filter",
			Description: "Clear current filter",
			Category:    "Filters",
			Action: func(m *Model) tea.Cmd {
				m.clearSearch()
				return nil
			},
		},
		// View
		{
			Name:        "markdown",
			Description: "Toggle markdown rendering",
			Category:    "View",
			Action: func(m *Model) tea.Cmd {
				m.detailRenderMarkdown = !m.detailRenderMarkdown
				return nil
			},
		},
		{
			Name:        "view compact",
			Description: "Compact view density",
			Category:    "View",
			Action: func(m *Model) tea.Cmd {
				m.viewDensity = "compact"
				return nil
			},
		},
		{
			Name:        "view comfortable",
			Description: "Comfortable view density",
			Category:    "View",
			Action: func(m *Model) tea.Cmd {
				m.viewDensity = "comfortable"
				return nil
			},
		},
		{
			Name:        "view expanded",
			Description: "Expanded view density",
			Category:    "View",
			Action: func(m *Model) tea.Cmd {
				m.viewDensity = "expanded"
				return nil
			},
		},
		// Themes
		{
			Name:        "theme dracula",
			Description: "Switch to Dracula theme",
			Category:    "Themes",
			Action:      makeThemeAction("dracula"),
		},
		{
			Name:        "theme nord",
			Description: "Switch to Nord theme",
			Category:    "Themes",
			Action:      makeThemeAction("nord"),
		},
		{
			Name:        "theme gruvbox",
			Description: "Switch to Gruvbox theme",
			Category:    "Themes",
			Action:      makeThemeAction("gruvbox"),
		},
		{
			Name:        "theme one-dark",
			Description: "Switch to One Dark theme",
			Category:    "Themes",
			Action:      makeThemeAction("one-dark"),
		},
		{
			Name:        "theme tokyo-night",
			Description: "Switch to Tokyo Night theme",
			Category:    "Themes",
			Action:      makeThemeAction("tokyo-night"),
		},
		{
			Name:        "theme catppuccin-mocha",
			Description: "Switch to Catppuccin Mocha theme",
			Category:    "Themes",
			Action:      makeThemeAction("catppuccin-mocha"),
		},
		{
			Name:        "theme monokai",
			Description: "Switch to Monokai theme",
			Category:    "Themes",
			Action:      makeThemeAction("monokai"),
		},
		{
			Name:        "theme solarized-dark",
			Description: "Switch to Solarized Dark theme",
			Category:    "Themes",
			Action:      makeThemeAction("solarized-dark"),
		},
		{
			Name:        "theme solarized-light",
			Description: "Switch to Solarized Light theme",
			Category:    "Themes",
			Action:      makeThemeAction("solarized-light"),
		},
		{
			Name:        "theme high-contrast-dark",
			Description: "Switch to High Contrast Dark theme",
			Category:    "Themes",
			Action:      makeThemeAction("high-contrast-dark"),
		},
		{
			Name:        "theme high-contrast-light",
			Description: "Switch to High Contrast Light theme",
			Category:    "Themes",
			Action:      makeThemeAction("high-contrast-light"),
		},
		// Export
		{
			Name:        "export markdown",
			Description: "Export current memo as markdown",
			Category:    "Export",
			Action: func(m *Model) tea.Cmd {
				return m.exportCurrentMemo("markdown")
			},
		},
		{
			Name:        "export json",
			Description: "Export current memo as JSON",
			Category:    "Export",
			Action: func(m *Model) tea.Cmd {
				return m.exportCurrentMemo("json")
			},
		},
		{
			Name:        "export all",
			Description: "Export all memos",
			Category:    "Export",
			Action: func(m *Model) tea.Cmd {
				return m.exportAllMemos()
			},
		},
		// System
		// Editor mode
		{
			Name:        "editor vim",
			Description: "Use vim-style editor",
			Category:    "Editor",
			Action: func(m *Model) tea.Cmd {
				m.editorMode = "vim"
				m.settings.Editor.Mode = "vim"
				m.statusMessage = "Editor mode: vim"
				return nil
			},
		},
		{
			Name:        "editor simple",
			Description: "Use simple text editor",
			Category:    "Editor",
			Action: func(m *Model) tea.Cmd {
				m.editorMode = "simple"
				m.settings.Editor.Mode = "simple"
				m.statusMessage = "Editor mode: simple"
				return nil
			},
		},
		{
			Name:        "help",
			Description: "Show help",
			Category:    "System",
			Action: func(m *Model) tea.Cmd {
				m.showingHelp = true
				return nil
			},
		},
		{
			Name:        "quit",
			Description: "Exit Memento",
			Category:    "System",
			Action: func(m *Model) tea.Cmd {
				return tea.Quit
			},
		},
	}
}

func makeThemeAction(themeName string) func(m *Model) tea.Cmd {
	return func(m *Model) tea.Cmd {
		m.settings.Theme = themeName
		ApplyTheme(m.settings)
		m.statusMessage = "Theme: " + themeName
		return nil
	}
}

// FilterCommands filters commands based on query using fuzzy matching.
func FilterCommands(commands []Command, query string) []Command {
	if query == "" {
		return commands
	}

	// Remove leading ":" from query if present
	query = strings.TrimPrefix(query, ":")

	source := commandSource{commands: commands}
	matches := fuzzy.FindFrom(query, source)

	result := make([]Command, 0, len(matches))
	for _, match := range matches {
		result = append(result, commands[match.Index])
	}
	return result
}
