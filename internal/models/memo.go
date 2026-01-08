package models

import (
	"strings"
	"time"

	"github.com/tlwhittaker/memento/internal/api"
)

type Memo struct {
	Name        string
	UID         string
	Content     string
	CreateTime  time.Time
	UpdateTime  time.Time
	DisplayTime time.Time
	RowStatus   string
	Visibility  string
	Pinned      bool
	Tags        []string
}

func (m *Memo) IsArchived() bool {
	return m.RowStatus == "ARCHIVED"
}

func (m *Memo) IsPublic() bool {
	return m.Visibility == "PUBLIC"
}

func (m *Memo) IsProtected() bool {
	return m.Visibility == "PROTECTED"
}

func (m *Memo) VisibilityIcon() string {
	switch m.Visibility {
	case "PUBLIC":
		return "🌐"
	case "PROTECTED":
		return "👥"
	default:
		return "🔒"
	}
}

func (m *Memo) ShortID() string {
	parts := strings.Split(m.Name, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return m.Name
}

func (m *Memo) Preview(maxLen int) string {
	content := strings.TrimSpace(m.Content)
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\r", "")

	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen-3] + "..."
}

func (m *Memo) Title() string {
	lines := strings.SplitN(m.Content, "\n", 2)
	title := strings.TrimSpace(lines[0])
	if title == "" {
		return "(empty)"
	}
	return title
}

func (m *Memo) FormattedDate() string {
	return m.DisplayTime.Format("Jan 02, 2006 3:04 PM")
}

func (m *Memo) ShortDate() string {
	now := time.Now()
	if m.DisplayTime.Year() == now.Year() {
		if m.DisplayTime.YearDay() == now.YearDay() {
			return m.DisplayTime.Format("3:04 PM")
		}
		return m.DisplayTime.Format("Jan 02")
	}
	return m.DisplayTime.Format("Jan 02, 2006")
}

func FromAPI(m api.Memo) Memo {
	return Memo{
		Name:        m.Name,
		UID:         m.UID,
		Content:     m.Content,
		CreateTime:  m.CreateTime,
		UpdateTime:  m.UpdateTime,
		DisplayTime: m.DisplayTime,
		RowStatus:   m.RowStatus,
		Visibility:  m.Visibility,
		Pinned:      m.Pinned,
		Tags:        m.Tags,
	}
}

func FromAPIList(memos []api.Memo) []Memo {
	result := make([]Memo, len(memos))
	for i, m := range memos {
		result[i] = FromAPI(m)
	}
	return result
}
