package ui

import (
	"github.com/charmbracelet/glamour"
)

var markdownRenderer *glamour.TermRenderer

// initMarkdownRenderer creates a glamour renderer with the current theme.
func initMarkdownRenderer(width int) error {
	// Create a renderer matching the current theme style
	style := getGlamourStyle()

	r, err := glamour.NewTermRenderer(
		glamour.WithStylePath(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return err
	}
	markdownRenderer = r
	return nil
}

// RenderMarkdown renders markdown content to styled terminal output.
func RenderMarkdown(content string, width int) (string, error) {
	// Initialize or update renderer if needed
	if markdownRenderer == nil {
		if err := initMarkdownRenderer(width); err != nil {
			return content, err
		}
	}

	rendered, err := markdownRenderer.Render(content)
	if err != nil {
		return content, err
	}
	return rendered, nil
}

// getGlamourStyle returns a glamour style name based on current theme.
func getGlamourStyle() string {
	// Map themes to glamour styles
	themeName := CurrentTheme.Name

	switch themeName {
	case "Solarized Light", "High Contrast Light":
		return "light"
	case "Dracula":
		return "dracula"
	case "Nord":
		return "nord"
	case "Tokyo Night":
		return "tokyo-night"
	default:
		return "dark"
	}
}

// ResetMarkdownRenderer clears the renderer (call when theme changes).
func ResetMarkdownRenderer() {
	markdownRenderer = nil
}
