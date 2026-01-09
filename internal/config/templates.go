package config

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Template represents a loaded template.
type Template struct {
	Name    string // Display name (filename without .md extension)
	Path    string // Full path to the template file
	Content string // Template content
}

// TemplatesDir returns the path to the templates directory.
func TemplatesDir() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "templates"), nil
}

// LoadTemplates discovers and loads all .md templates from the templates directory.
func LoadTemplates() ([]Template, error) {
	dir, err := TemplatesDir()
	if err != nil {
		return nil, err
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return []Template{}, nil // No templates directory = no templates
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var templates []Template
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".md") {
			continue
		}

		path := filepath.Join(dir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			continue // Skip files that can't be read
		}

		// Skip empty templates
		trimmed := strings.TrimSpace(string(content))
		if trimmed == "" {
			continue
		}

		templates = append(templates, Template{
			Name:    strings.TrimSuffix(name, filepath.Ext(name)),
			Path:    path,
			Content: string(content),
		})
	}

	// Sort templates alphabetically by name
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Name < templates[j].Name
	})

	return templates, nil
}
