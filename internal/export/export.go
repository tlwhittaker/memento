package export

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tlwhittaker/memento/internal/config"
	"github.com/tlwhittaker/memento/internal/models"
)

// ExportFormat represents the export file format.
type ExportFormat string

const (
	FormatMarkdown ExportFormat = "markdown"
	FormatJSON     ExportFormat = "json"
	FormatText     ExportFormat = "text"
)

// ExportMemo exports a single memo to a file.
func ExportMemo(memo models.Memo, format ExportFormat, dir string) (string, error) {
	if dir == "" {
		var err error
		dir, err = getExportDir()
		if err != nil {
			return "", err
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create export directory: %w", err)
	}

	// Generate filename
	filename := generateFilename(memo, format)
	filepath := filepath.Join(dir, filename)

	// Generate content
	var content string
	switch format {
	case FormatMarkdown:
		content = toMarkdown(memo)
	case FormatJSON:
		data, err := json.MarshalIndent(memo, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal memo: %w", err)
		}
		content = string(data)
	case FormatText:
		content = toText(memo)
	default:
		content = memo.Content
	}

	// Write file
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filepath, nil
}

// ExportMemos exports multiple memos to a directory.
func ExportMemos(memos []models.Memo, format ExportFormat) (string, error) {
	exportDir, err := getExportDir()
	if err != nil {
		return "", err
	}

	// Create timestamped subdirectory
	timestamp := time.Now().Format("2006-01-02_150405")
	dir := filepath.Join(exportDir, fmt.Sprintf("export_%s", timestamp))

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create export directory: %w", err)
	}

	for _, memo := range memos {
		if _, err := ExportMemo(memo, format, dir); err != nil {
			return dir, err
		}
	}

	return dir, nil
}

func getExportDir() (string, error) {
	dataDir, err := config.DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "exports"), nil
}

func generateFilename(memo models.Memo, format ExportFormat) string {
	// Use first line of content as base name, sanitized
	title := memo.Title()
	if len(title) > 50 {
		title = title[:50]
	}

	// Sanitize filename
	title = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
			return '_'
		}
		return r
	}, title)

	title = strings.TrimSpace(title)
	if title == "" {
		title = "memo"
	}

	ext := ".txt"
	switch format {
	case FormatMarkdown:
		ext = ".md"
	case FormatJSON:
		ext = ".json"
	}

	return fmt.Sprintf("%s_%s%s", memo.ShortID(), title, ext)
}

func toMarkdown(memo models.Memo) string {
	var sb strings.Builder

	// Frontmatter
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("id: %s\n", memo.ShortID()))
	sb.WriteString(fmt.Sprintf("created: %s\n", memo.CreateTime.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("updated: %s\n", memo.UpdateTime.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("visibility: %s\n", memo.Visibility))
	if memo.Pinned {
		sb.WriteString("pinned: true\n")
	}
	if len(memo.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(memo.Tags, ", ")))
	}
	sb.WriteString("---\n\n")

	// Content
	sb.WriteString(memo.Content)
	sb.WriteString("\n")

	return sb.String()
}

func toText(memo models.Memo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Memo #%s\n", memo.ShortID()))
	sb.WriteString(fmt.Sprintf("Created: %s\n", memo.FormattedDate()))
	sb.WriteString(fmt.Sprintf("Visibility: %s\n", memo.Visibility))
	if memo.Pinned {
		sb.WriteString("Pinned: yes\n")
	}
	sb.WriteString(strings.Repeat("-", 40))
	sb.WriteString("\n\n")
	sb.WriteString(memo.Content)
	sb.WriteString("\n")

	return sb.String()
}
