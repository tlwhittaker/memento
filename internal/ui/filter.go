package ui

import (
	"regexp"
	"strings"
	"time"

	"github.com/sahilm/fuzzy"
	"github.com/tlwhittaker/memento/internal/models"
)

// DateRange represents a date range filter.
type DateRange struct {
	Start time.Time
	End   time.Time
}

// Filter represents parsed search/filter criteria.
type Filter struct {
	Text       string     // Fuzzy text search
	Tags       []string   // #tag filters
	Visibility string     // v:public, v:private, v:protected
	Pinned     *bool      // pinned status filter
	Archived   *bool      // archived status filter
	DateRange  *DateRange // Date range filter
}

var (
	tagPattern        = regexp.MustCompile(`#(\w+)`)
	visibilityPattern = regexp.MustCompile(`v:(public|private|protected)`)
	datePattern       = regexp.MustCompile(`date:(\S+)`)
)

// ParseFilter parses a search query into a Filter struct.
func ParseFilter(query string) Filter {
	filter := Filter{}

	// Extract tags (#tag or t:tag)
	tagMatches := tagPattern.FindAllStringSubmatch(query, -1)
	for _, match := range tagMatches {
		if len(match) > 1 {
			filter.Tags = append(filter.Tags, match[1])
		}
	}
	query = tagPattern.ReplaceAllString(query, "")

	// Also handle t:tag format
	tTagPattern := regexp.MustCompile(`t:(\w+)`)
	tTagMatches := tTagPattern.FindAllStringSubmatch(query, -1)
	for _, match := range tTagMatches {
		if len(match) > 1 {
			filter.Tags = append(filter.Tags, match[1])
		}
	}
	query = tTagPattern.ReplaceAllString(query, "")

	// Extract visibility
	visMatches := visibilityPattern.FindStringSubmatch(query)
	if len(visMatches) > 1 {
		filter.Visibility = strings.ToUpper(visMatches[1])
	}
	query = visibilityPattern.ReplaceAllString(query, "")

	// Extract date range
	dateMatches := datePattern.FindStringSubmatch(query)
	if len(dateMatches) > 1 {
		filter.DateRange = parseDateRange(dateMatches[1])
	}
	query = datePattern.ReplaceAllString(query, "")

	// Check for pinned/unpinned
	if strings.Contains(query, "pinned") {
		pinned := true
		filter.Pinned = &pinned
		query = strings.ReplaceAll(query, "pinned", "")
	} else if strings.Contains(query, "!pinned") || strings.Contains(query, "unpinned") {
		pinned := false
		filter.Pinned = &pinned
		query = strings.ReplaceAll(query, "!pinned", "")
		query = strings.ReplaceAll(query, "unpinned", "")
	}

	// Check for archived/unarchived
	if strings.Contains(query, "archived") {
		archived := true
		filter.Archived = &archived
		query = strings.ReplaceAll(query, "archived", "")
	} else if strings.Contains(query, "!archived") || strings.Contains(query, "unarchived") {
		archived := false
		filter.Archived = &archived
		query = strings.ReplaceAll(query, "!archived", "")
		query = strings.ReplaceAll(query, "unarchived", "")
	}

	// Remaining text is the fuzzy search query
	filter.Text = strings.TrimSpace(query)

	return filter
}

// parseDateRange parses date range strings like "today", "this-week", "2024-01"
func parseDateRange(dateStr string) *DateRange {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	switch dateStr {
	case "today":
		return &DateRange{
			Start: today,
			End:   today.Add(24 * time.Hour),
		}
	case "yesterday":
		yesterday := today.Add(-24 * time.Hour)
		return &DateRange{
			Start: yesterday,
			End:   today,
		}
	case "this-week":
		// Start of week (Sunday)
		weekday := int(now.Weekday())
		startOfWeek := today.Add(-time.Duration(weekday) * 24 * time.Hour)
		return &DateRange{
			Start: startOfWeek,
			End:   startOfWeek.Add(7 * 24 * time.Hour),
		}
	case "last-week":
		weekday := int(now.Weekday())
		startOfThisWeek := today.Add(-time.Duration(weekday) * 24 * time.Hour)
		startOfLastWeek := startOfThisWeek.Add(-7 * 24 * time.Hour)
		return &DateRange{
			Start: startOfLastWeek,
			End:   startOfThisWeek,
		}
	case "this-month":
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		startOfNextMonth := startOfMonth.AddDate(0, 1, 0)
		return &DateRange{
			Start: startOfMonth,
			End:   startOfNextMonth,
		}
	case "last-month":
		startOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		startOfLastMonth := startOfThisMonth.AddDate(0, -1, 0)
		return &DateRange{
			Start: startOfLastMonth,
			End:   startOfThisMonth,
		}
	default:
		// Try to parse as YYYY-MM format
		if len(dateStr) == 7 && dateStr[4] == '-' {
			t, err := time.Parse("2006-01", dateStr)
			if err == nil {
				return &DateRange{
					Start: t,
					End:   t.AddDate(0, 1, 0),
				}
			}
		}
		// Try to parse as YYYY-MM-DD format
		if len(dateStr) == 10 {
			t, err := time.Parse("2006-01-02", dateStr)
			if err == nil {
				return &DateRange{
					Start: t,
					End:   t.Add(24 * time.Hour),
				}
			}
		}
	}

	return nil
}

// ApplyFilter applies a filter to a list of memos.
func ApplyFilter(memos []models.Memo, filter Filter) []models.Memo {
	result := make([]models.Memo, 0, len(memos))

	for _, memo := range memos {
		if matchesFilter(memo, filter) {
			result = append(result, memo)
		}
	}

	// If there's text to fuzzy match, apply fuzzy filtering
	if filter.Text != "" && len(result) > 0 {
		source := memoSearcher{memos: result}
		matches := fuzzy.FindFrom(filter.Text, source)
		filtered := make([]models.Memo, 0, len(matches))
		for _, match := range matches {
			filtered = append(filtered, result[match.Index])
		}
		result = filtered
	}

	return result
}

// matchesFilter checks if a memo matches the non-text filter criteria.
func matchesFilter(memo models.Memo, filter Filter) bool {
	// Check pinned
	if filter.Pinned != nil && memo.Pinned != *filter.Pinned {
		return false
	}

	// Check archived
	if filter.Archived != nil && memo.IsArchived() != *filter.Archived {
		return false
	}

	// Check visibility
	if filter.Visibility != "" && memo.Visibility != filter.Visibility {
		return false
	}

	// Check tags
	if len(filter.Tags) > 0 {
		memoTags := make(map[string]bool)
		for _, tag := range memo.Tags {
			memoTags[strings.ToLower(tag)] = true
		}
		// Also extract tags from content
		contentTags := tagPattern.FindAllStringSubmatch(memo.Content, -1)
		for _, match := range contentTags {
			if len(match) > 1 {
				memoTags[strings.ToLower(match[1])] = true
			}
		}

		for _, filterTag := range filter.Tags {
			if !memoTags[strings.ToLower(filterTag)] {
				return false
			}
		}
	}

	// Check date range
	if filter.DateRange != nil {
		if memo.DisplayTime.Before(filter.DateRange.Start) ||
			memo.DisplayTime.After(filter.DateRange.End) {
			return false
		}
	}

	return true
}

// ExtractTags extracts all unique tags from memos.
func ExtractTags(memos []models.Memo) []TagInfo {
	tagCounts := make(map[string]int)

	for _, memo := range memos {
		// From memo.Tags field
		for _, tag := range memo.Tags {
			tagCounts[strings.ToLower(tag)]++
		}
		// From content (hashtags)
		matches := tagPattern.FindAllStringSubmatch(memo.Content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				tagCounts[strings.ToLower(match[1])]++
			}
		}
	}

	tags := make([]TagInfo, 0, len(tagCounts))
	for name, count := range tagCounts {
		tags = append(tags, TagInfo{
			Name:  name,
			Count: count,
		})
	}

	return tags
}

// TagInfo holds information about a tag.
type TagInfo struct {
	Name  string
	Count int
}
