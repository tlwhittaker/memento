package ui

import "github.com/charmbracelet/lipgloss"

// Theme holds all colors for a UI theme.
type Theme struct {
	Name       string
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Error      lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Muted      lipgloss.Color
	Text       lipgloss.Color
	Background lipgloss.Color
	Selected   lipgloss.Color
}

// Built-in themes
var Themes = map[string]Theme{
	"dracula": {
		Name:       "Dracula",
		Primary:    lipgloss.Color("#BD93F9"),
		Secondary:  lipgloss.Color("#8BE9FD"),
		Error:      lipgloss.Color("#FF5555"),
		Success:    lipgloss.Color("#50FA7B"),
		Warning:    lipgloss.Color("#FFB86C"),
		Muted:      lipgloss.Color("#6272A4"),
		Text:       lipgloss.Color("#F8F8F2"),
		Background: lipgloss.Color("#282A36"),
		Selected:   lipgloss.Color("#44475A"),
	},
	"nord": {
		Name:       "Nord",
		Primary:    lipgloss.Color("#88C0D0"),
		Secondary:  lipgloss.Color("#81A1C1"),
		Error:      lipgloss.Color("#BF616A"),
		Success:    lipgloss.Color("#A3BE8C"),
		Warning:    lipgloss.Color("#EBCB8B"),
		Muted:      lipgloss.Color("#4C566A"),
		Text:       lipgloss.Color("#ECEFF4"),
		Background: lipgloss.Color("#2E3440"),
		Selected:   lipgloss.Color("#3B4252"),
	},
	"solarized-dark": {
		Name:       "Solarized Dark",
		Primary:    lipgloss.Color("#268BD2"),
		Secondary:  lipgloss.Color("#2AA198"),
		Error:      lipgloss.Color("#DC322F"),
		Success:    lipgloss.Color("#859900"),
		Warning:    lipgloss.Color("#B58900"),
		Muted:      lipgloss.Color("#586E75"),
		Text:       lipgloss.Color("#839496"),
		Background: lipgloss.Color("#002B36"),
		Selected:   lipgloss.Color("#073642"),
	},
	"solarized-light": {
		Name:       "Solarized Light",
		Primary:    lipgloss.Color("#268BD2"),
		Secondary:  lipgloss.Color("#2AA198"),
		Error:      lipgloss.Color("#DC322F"),
		Success:    lipgloss.Color("#859900"),
		Warning:    lipgloss.Color("#B58900"),
		Muted:      lipgloss.Color("#93A1A1"),
		Text:       lipgloss.Color("#657B83"),
		Background: lipgloss.Color("#FDF6E3"),
		Selected:   lipgloss.Color("#EEE8D5"),
	},
	"gruvbox": {
		Name:       "Gruvbox",
		Primary:    lipgloss.Color("#83A598"),
		Secondary:  lipgloss.Color("#8EC07C"),
		Error:      lipgloss.Color("#FB4934"),
		Success:    lipgloss.Color("#B8BB26"),
		Warning:    lipgloss.Color("#FABD2F"),
		Muted:      lipgloss.Color("#928374"),
		Text:       lipgloss.Color("#EBDBB2"),
		Background: lipgloss.Color("#282828"),
		Selected:   lipgloss.Color("#3C3836"),
	},
	"one-dark": {
		Name:       "One Dark",
		Primary:    lipgloss.Color("#61AFEF"),
		Secondary:  lipgloss.Color("#56B6C2"),
		Error:      lipgloss.Color("#E06C75"),
		Success:    lipgloss.Color("#98C379"),
		Warning:    lipgloss.Color("#E5C07B"),
		Muted:      lipgloss.Color("#5C6370"),
		Text:       lipgloss.Color("#ABB2BF"),
		Background: lipgloss.Color("#282C34"),
		Selected:   lipgloss.Color("#3E4451"),
	},
	"tokyo-night": {
		Name:       "Tokyo Night",
		Primary:    lipgloss.Color("#7AA2F7"),
		Secondary:  lipgloss.Color("#7DCFFF"),
		Error:      lipgloss.Color("#F7768E"),
		Success:    lipgloss.Color("#9ECE6A"),
		Warning:    lipgloss.Color("#E0AF68"),
		Muted:      lipgloss.Color("#565F89"),
		Text:       lipgloss.Color("#C0CAF5"),
		Background: lipgloss.Color("#1A1B26"),
		Selected:   lipgloss.Color("#292E42"),
	},
	"catppuccin-mocha": {
		Name:       "Catppuccin Mocha",
		Primary:    lipgloss.Color("#CBA6F7"),
		Secondary:  lipgloss.Color("#89DCEB"),
		Error:      lipgloss.Color("#F38BA8"),
		Success:    lipgloss.Color("#A6E3A1"),
		Warning:    lipgloss.Color("#F9E2AF"),
		Muted:      lipgloss.Color("#6C7086"),
		Text:       lipgloss.Color("#CDD6F4"),
		Background: lipgloss.Color("#1E1E2E"),
		Selected:   lipgloss.Color("#313244"),
	},
	"monokai": {
		Name:       "Monokai",
		Primary:    lipgloss.Color("#66D9EF"),
		Secondary:  lipgloss.Color("#A6E22E"),
		Error:      lipgloss.Color("#F92672"),
		Success:    lipgloss.Color("#A6E22E"),
		Warning:    lipgloss.Color("#FD971F"),
		Muted:      lipgloss.Color("#75715E"),
		Text:       lipgloss.Color("#F8F8F2"),
		Background: lipgloss.Color("#272822"),
		Selected:   lipgloss.Color("#3E3D32"),
	},
	"high-contrast-dark": {
		Name:       "High Contrast Dark",
		Primary:    lipgloss.Color("#FFFFFF"),
		Secondary:  lipgloss.Color("#00FFFF"),
		Error:      lipgloss.Color("#FF0000"),
		Success:    lipgloss.Color("#00FF00"),
		Warning:    lipgloss.Color("#FFFF00"),
		Muted:      lipgloss.Color("#808080"),
		Text:       lipgloss.Color("#FFFFFF"),
		Background: lipgloss.Color("#000000"),
		Selected:   lipgloss.Color("#333333"),
	},
	"high-contrast-light": {
		Name:       "High Contrast Light",
		Primary:    lipgloss.Color("#000000"),
		Secondary:  lipgloss.Color("#0000AA"),
		Error:      lipgloss.Color("#AA0000"),
		Success:    lipgloss.Color("#006600"),
		Warning:    lipgloss.Color("#AA6600"),
		Muted:      lipgloss.Color("#666666"),
		Text:       lipgloss.Color("#000000"),
		Background: lipgloss.Color("#FFFFFF"),
		Selected:   lipgloss.Color("#DDDDDD"),
	},
}

// GetTheme returns a theme by name, falling back to dracula if not found.
func GetTheme(name string) Theme {
	if theme, ok := Themes[name]; ok {
		return theme
	}
	return Themes["dracula"]
}

// CustomTheme creates a theme from custom color values.
func CustomTheme(primary, secondary, errorC, success, warning, muted, text, background, selected string) Theme {
	return Theme{
		Name:       "Custom",
		Primary:    lipgloss.Color(primary),
		Secondary:  lipgloss.Color(secondary),
		Error:      lipgloss.Color(errorC),
		Success:    lipgloss.Color(success),
		Warning:    lipgloss.Color(warning),
		Muted:      lipgloss.Color(muted),
		Text:       lipgloss.Color(text),
		Background: lipgloss.Color(background),
		Selected:   lipgloss.Color(selected),
	}
}

// ThemeNames returns a list of available theme names.
func ThemeNames() []string {
	names := make([]string, 0, len(Themes))
	for name := range Themes {
		names = append(names, name)
	}
	return names
}
