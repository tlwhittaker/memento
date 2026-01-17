package models

import (
	"fmt"
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
	Resources   []Resource
}

// Resource represents an attachment on a memo.
type Resource struct {
	Name         string
	UID          string
	Filename     string
	Type         string
	Size         int64
	CreateTime   time.Time
	ExternalLink string
}

// HasResources returns true if the memo has any resources attached.
func (m *Memo) HasResources() bool {
	return len(m.Resources) > 0
}

// ResourceCount returns the number of resources attached.
func (m *Memo) ResourceCount() int {
	return len(m.Resources)
}

// FormatFileSize returns a human-readable file size string.
func FormatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.1f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// ResourceIcon returns an icon based on the resource type.
func (r *Resource) Icon() string {
	switch {
	case strings.HasPrefix(r.Type, "image/"):
		return "[img]"
	case strings.HasPrefix(r.Type, "video/"):
		return "[vid]"
	case strings.HasPrefix(r.Type, "audio/"):
		return "[aud]"
	case r.Type == "application/pdf":
		return "[pdf]"
	default:
		return "[file]"
	}
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
	if maxLen <= 0 {
		return ""
	}
	if maxLen <= 3 {
		return "..."[:maxLen]
	}

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

func (m *Memo) FormattedDateIn(tz *time.Location) string {
	return m.DisplayTime.In(tz).Format("Jan 02, 2006 3:04 PM")
}

func (m *Memo) ShortDate() string {
	return m.ShortDateIn(time.Local)
}

func (m *Memo) ShortDateIn(tz *time.Location) string {
	localTime := m.DisplayTime.In(tz)
	now := time.Now().In(tz)
	if localTime.Year() == now.Year() {
		if localTime.YearDay() == now.YearDay() {
			return localTime.Format("3:04 PM")
		}
		return localTime.Format("Jan 02")
	}
	return localTime.Format("Jan 02, 2006")
}

// RelativeDate returns a human-readable relative date string.
func (m *Memo) RelativeDate() string {
	return m.RelativeDateIn(time.Local)
}

// RelativeDateIn returns a human-readable relative date string in the given timezone.
func (m *Memo) RelativeDateIn(tz *time.Location) string {
	now := time.Now().In(tz)
	t := m.DisplayTime.In(tz)
	diff := now.Sub(t)

	// Future dates
	if diff < 0 {
		return t.Format("Jan 02, 2006")
	}

	// Less than a minute
	if diff < time.Minute {
		return "just now"
	}

	// Less than an hour
	if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	}

	// Less than a day
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	// Yesterday
	yesterday := now.AddDate(0, 0, -1)
	if t.Year() == yesterday.Year() && t.YearDay() == yesterday.YearDay() {
		return "yesterday"
	}

	// Less than a week
	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	// More than a week - show date
	if t.Year() == now.Year() {
		return t.Format("Jan 02")
	}
	return t.Format("Jan 02, 2006")
}

func FromAPI(m api.Memo) Memo {
	// Convert resources
	resources := make([]Resource, len(m.Resources))
	for i, r := range m.Resources {
		resources[i] = Resource{
			Name:         r.Name,
			UID:          r.UID,
			Filename:     r.Filename,
			Type:         r.Type,
			Size:         r.Size,
			CreateTime:   r.CreateTime,
			ExternalLink: r.ExternalLink,
		}
	}

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
		Resources:   resources,
	}
}

func FromAPIList(memos []api.Memo) []Memo {
	result := make([]Memo, len(memos))
	for i, m := range memos {
		result[i] = FromAPI(m)
	}
	return result
}
