package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/tlwhittaker/memento/internal/config"
)

// CurrentTheme holds the active theme colors
var CurrentTheme Theme

// Colors - these are updated when ApplyTheme is called
var (
	PrimaryColor    lipgloss.Color
	SecondaryColor  lipgloss.Color
	ErrorColor      lipgloss.Color
	SuccessColor    lipgloss.Color
	WarningColor    lipgloss.Color
	MutedColor      lipgloss.Color
	TextColor       lipgloss.Color
	BgColor         lipgloss.Color
	SelectedBgColor lipgloss.Color
)

// ApplyTheme sets the current theme and updates all color variables.
func ApplyTheme(settings *config.Settings) {
	var theme Theme

	if settings.Theme == "custom" {
		theme = CustomTheme(
			settings.Colors.Primary,
			settings.Colors.Secondary,
			settings.Colors.Error,
			settings.Colors.Success,
			settings.Colors.Warning,
			settings.Colors.Muted,
			settings.Colors.Text,
			settings.Colors.Background,
			settings.Colors.Selected,
		)
	} else {
		theme = GetTheme(settings.Theme)
	}

	CurrentTheme = theme
	PrimaryColor = theme.Primary
	SecondaryColor = theme.Secondary
	ErrorColor = theme.Error
	SuccessColor = theme.Success
	WarningColor = theme.Warning
	MutedColor = theme.Muted
	TextColor = theme.Text
	BgColor = theme.Background
	SelectedBgColor = theme.Selected

	// Rebuild all styles with new colors
	rebuildStyles()
}

func init() {
	theme := GetTheme("dracula")
	CurrentTheme = theme
	PrimaryColor = theme.Primary
	SecondaryColor = theme.Secondary
	ErrorColor = theme.Error
	SuccessColor = theme.Success
	WarningColor = theme.Warning
	MutedColor = theme.Muted
	TextColor = theme.Text
	BgColor = theme.Background
	SelectedBgColor = theme.Selected
	rebuildStyles()
}

// Common styles
var (
	TitleStyle    lipgloss.Style
	SubtitleStyle lipgloss.Style
	HelpStyle     lipgloss.Style
	MutedStyle    lipgloss.Style
	ErrorStyle    lipgloss.Style
	SuccessStyle  lipgloss.Style
	WarningStyle  lipgloss.Style
)

// Box styles for consistent borders
var (
	BoxStyle        lipgloss.Style
	ContentBoxStyle lipgloss.Style
	TextAreaStyle   lipgloss.Style
)

// Memo-specific styles
var (
	MemoTitleStyle           lipgloss.Style
	MemoDateStyle            lipgloss.Style
	MemoPreviewStyle         lipgloss.Style
	MemoSelectedStyle        lipgloss.Style
	MemoSelectedBgStyle      lipgloss.Style
	MemoSelectedPreviewStyle lipgloss.Style
	MemoIDStyle              lipgloss.Style
	PinnedStyle              lipgloss.Style
)

// StatusBar styles
var (
	StatusBarStyle  lipgloss.Style
	StatusKeyStyle  lipgloss.Style
	StatusSeparator lipgloss.Style
)

// Header styles
var (
	HeaderStyle    lipgloss.Style
	HeaderBoxStyle lipgloss.Style
)

// Confirmation dialog styles
var (
	DialogStyle             lipgloss.Style
	DialogTitleStyle        lipgloss.Style
	UnsavedDialogStyle      lipgloss.Style
	UnsavedDialogTitleStyle lipgloss.Style
)

// Template picker styles
var (
	TemplatePickerStyle      lipgloss.Style
	TemplatePickerTitleStyle lipgloss.Style
)

// Cursor style
var CursorStyle lipgloss.Style

// Vim mode indicator styles
var (
	NormalModeStyle  lipgloss.Style
	InsertModeStyle  lipgloss.Style
	VisualModeStyle  lipgloss.Style
	CommandModeStyle lipgloss.Style
)

// rebuildStyles recreates all styles with current theme colors.
func rebuildStyles() {
	// Common styles
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Italic(true)

	HelpStyle = lipgloss.NewStyle().
		Foreground(MutedColor)

	MutedStyle = lipgloss.NewStyle().
		Foreground(MutedColor)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true)

	SuccessStyle = lipgloss.NewStyle().
		Foreground(SuccessColor)

	WarningStyle = lipgloss.NewStyle().
		Foreground(WarningColor)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor)

	ContentBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(MutedColor)

	TextAreaStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(MutedColor).
		Padding(0, 1)

	// Memo styles
	MemoTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(TextColor)

	MemoDateStyle = lipgloss.NewStyle().
		Foreground(MutedColor)

	MemoPreviewStyle = lipgloss.NewStyle().
		Foreground(TextColor)

	MemoSelectedStyle = lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Bold(true)

	MemoSelectedBgStyle = lipgloss.NewStyle().
		Background(SelectedBgColor).
		Foreground(SecondaryColor).
		Bold(true)

	MemoSelectedPreviewStyle = lipgloss.NewStyle().
		Background(SelectedBgColor).
		Foreground(SecondaryColor)

	MemoIDStyle = lipgloss.NewStyle().
		Foreground(MutedColor)

	PinnedStyle = lipgloss.NewStyle().
		Foreground(WarningColor)

	// StatusBar styles
	StatusBarStyle = lipgloss.NewStyle().
		Foreground(MutedColor)

	StatusKeyStyle = lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Bold(true)

	StatusSeparator = lipgloss.NewStyle().
		Foreground(MutedColor).
		SetString(" │ ")

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		Padding(0, 1)

	HeaderBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor)

	// Dialog styles
	DialogStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(WarningColor).
		Padding(1, 2).
		Align(lipgloss.Center)

	DialogTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(WarningColor)

	UnsavedDialogStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SecondaryColor).
		Padding(1, 2).
		Align(lipgloss.Center)

	UnsavedDialogTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(SecondaryColor)

	// Template picker styles
	TemplatePickerStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2).
		Align(lipgloss.Left)

	TemplatePickerTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor)

	// Cursor style
	CursorStyle = lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Bold(true)

	// Vim mode styles
	NormalModeStyle = lipgloss.NewStyle().
		Background(PrimaryColor).
		Foreground(BgColor).
		Bold(true).
		Padding(0, 1)

	InsertModeStyle = lipgloss.NewStyle().
		Background(SuccessColor).
		Foreground(BgColor).
		Bold(true).
		Padding(0, 1)

	VisualModeStyle = lipgloss.NewStyle().
		Background(WarningColor).
		Foreground(BgColor).
		Bold(true).
		Padding(0, 1)

	CommandModeStyle = lipgloss.NewStyle().
		Background(SecondaryColor).
		Foreground(BgColor).
		Bold(true).
		Padding(0, 1)
}
