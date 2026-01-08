package ui

import (
	"strings"
	"unicode"
)

// Vim movement and editing functions for the editor

// isWordChar returns true if the rune is part of a word (alphanumeric or underscore).
func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// WordForward moves cursor to the start of the next word.
func WordForward(content string, cursor int) int {
	if cursor >= len(content) {
		return cursor
	}

	runes := []rune(content)
	pos := cursor

	// Skip current word if we're in one
	for pos < len(runes) && isWordChar(runes[pos]) {
		pos++
	}

	// Skip whitespace/punctuation
	for pos < len(runes) && !isWordChar(runes[pos]) {
		pos++
	}

	return pos
}

// WordBackward moves cursor to the start of the previous word.
func WordBackward(content string, cursor int) int {
	if cursor <= 0 {
		return 0
	}

	runes := []rune(content)
	pos := cursor - 1

	// Skip whitespace/punctuation before cursor
	for pos > 0 && !isWordChar(runes[pos]) {
		pos--
	}

	// Move to start of word
	for pos > 0 && isWordChar(runes[pos-1]) {
		pos--
	}

	return pos
}

// WordEnd moves cursor to the end of the current/next word.
func WordEnd(content string, cursor int) int {
	if cursor >= len(content)-1 {
		return len(content) - 1
	}

	runes := []rune(content)
	pos := cursor + 1

	// Skip whitespace/punctuation
	for pos < len(runes) && !isWordChar(runes[pos]) {
		pos++
	}

	// Move to end of word
	for pos < len(runes)-1 && isWordChar(runes[pos+1]) {
		pos++
	}

	return pos
}

// LineStart moves cursor to the start of the current line.
func LineStart(content string, cursor int) int {
	if cursor <= 0 {
		return 0
	}

	// Find the newline before cursor
	pos := cursor - 1
	for pos > 0 && content[pos] != '\n' {
		pos--
	}

	if content[pos] == '\n' {
		return pos + 1
	}
	return 0
}

// LineEnd moves cursor to the end of the current line.
func LineEnd(content string, cursor int) int {
	if cursor >= len(content) {
		return len(content)
	}

	pos := cursor
	for pos < len(content) && content[pos] != '\n' {
		pos++
	}

	return pos
}

// DocumentStart returns position 0 (start of document).
func DocumentStart() int {
	return 0
}

// DocumentEnd returns the last position in the document.
func DocumentEnd(content string) int {
	return len(content)
}

// FindNextChar finds the next occurrence of a character after cursor.
func FindNextChar(content string, cursor int, char rune) int {
	if cursor >= len(content)-1 {
		return cursor
	}

	runes := []rune(content)
	for i := cursor + 1; i < len(runes); i++ {
		if runes[i] == char {
			return i
		}
	}

	return cursor // Not found, stay in place
}

// FindPrevChar finds the previous occurrence of a character before cursor.
func FindPrevChar(content string, cursor int, char rune) int {
	if cursor <= 0 {
		return cursor
	}

	runes := []rune(content)
	for i := cursor - 1; i >= 0; i-- {
		if runes[i] == char {
			return i
		}
	}

	return cursor // Not found, stay in place
}

// FirstNonBlank returns the position of the first non-blank character on the current line.
func FirstNonBlank(content string, cursor int) int {
	lineStart := LineStart(content, cursor)
	pos := lineStart

	for pos < len(content) && content[pos] != '\n' && (content[pos] == ' ' || content[pos] == '\t') {
		pos++
	}

	return pos
}

// GetCurrentLine returns the content of the current line.
func GetCurrentLine(content string, cursor int) string {
	start := LineStart(content, cursor)
	end := LineEnd(content, cursor)
	return content[start:end]
}

// GetLineNumber returns the 0-indexed line number for the cursor position.
func GetLineNumber(content string, cursor int) int {
	line := 0
	for i := 0; i < cursor && i < len(content); i++ {
		if content[i] == '\n' {
			line++
		}
	}
	return line
}

// GetColumnNumber returns the 0-indexed column number for the cursor position.
func GetColumnNumber(content string, cursor int) int {
	lineStart := LineStart(content, cursor)
	return cursor - lineStart
}

// GoToLine moves cursor to the start of a specific line (0-indexed).
func GoToLine(content string, line int) int {
	if line <= 0 {
		return 0
	}

	pos := 0
	currentLine := 0

	for pos < len(content) {
		if currentLine == line {
			return pos
		}
		if content[pos] == '\n' {
			currentLine++
		}
		pos++
	}

	return pos // End of document if line doesn't exist
}

// DeleteLine removes the current line and returns the new content.
func DeleteLine(content string, cursor int) (string, int) {
	if content == "" {
		return "", 0
	}

	start := LineStart(content, cursor)
	end := LineEnd(content, cursor)

	// Include the newline character if present
	if end < len(content) && content[end] == '\n' {
		end++
	} else if start > 0 {
		// If deleting last line, remove preceding newline
		start--
	}

	newContent := content[:start] + content[end:]
	newCursor := start
	if newCursor > len(newContent) {
		newCursor = len(newContent)
	}

	return newContent, newCursor
}

// DeleteWord removes from cursor to end of word.
func DeleteWord(content string, cursor int) (string, int) {
	if cursor >= len(content) {
		return content, cursor
	}

	end := WordForward(content, cursor)
	newContent := content[:cursor] + content[end:]
	return newContent, cursor
}

// DeleteToLineEnd removes from cursor to end of line.
func DeleteToLineEnd(content string, cursor int) (string, int) {
	end := LineEnd(content, cursor)
	newContent := content[:cursor] + content[end:]
	return newContent, cursor
}

// YankLine returns the content of the current line (for clipboard).
func YankLine(content string, cursor int) string {
	return GetCurrentLine(content, cursor) + "\n"
}

// YankWord returns from cursor to end of word.
func YankWord(content string, cursor int) string {
	end := WordForward(content, cursor)
	return content[cursor:end]
}

// PasteBefore inserts text before the cursor.
func PasteBefore(content string, cursor int, text string) (string, int) {
	newContent := content[:cursor] + text + content[cursor:]
	return newContent, cursor + len(text)
}

// PasteAfter inserts text after the cursor.
func PasteAfter(content string, cursor int, text string) (string, int) {
	insertPos := cursor
	if cursor < len(content) {
		insertPos++
	}
	newContent := content[:insertPos] + text + content[insertPos:]
	return newContent, insertPos + len(text)
}

// InsertNewLineBelow inserts a new line below current and returns new cursor position.
func InsertNewLineBelow(content string, cursor int) (string, int) {
	end := LineEnd(content, cursor)
	newContent := content[:end] + "\n" + content[end:]
	return newContent, end + 1
}

// InsertNewLineAbove inserts a new line above current and returns new cursor position.
func InsertNewLineAbove(content string, cursor int) (string, int) {
	start := LineStart(content, cursor)
	newContent := content[:start] + "\n" + content[start:]
	return newContent, start
}

// ChangeWord deletes to end of word and returns new content (enters insert mode).
func ChangeWord(content string, cursor int) (string, int) {
	return DeleteWord(content, cursor)
}

// ChangeToLineEnd deletes to end of line and returns new content (enters insert mode).
func ChangeToLineEnd(content string, cursor int) (string, int) {
	return DeleteToLineEnd(content, cursor)
}

// GetSelectedText returns text between two positions (for visual mode).
func GetSelectedText(content string, start, end int) string {
	if start > end {
		start, end = end, start
	}
	if start < 0 {
		start = 0
	}
	if end > len(content) {
		end = len(content)
	}
	return content[start:end]
}

// DeleteSelection removes text between two positions.
func DeleteSelection(content string, start, end int) (string, int) {
	if start > end {
		start, end = end, start
	}
	if start < 0 {
		start = 0
	}
	if end > len(content) {
		end = len(content)
	}
	newContent := content[:start] + content[end:]
	return newContent, start
}

// MoveUp moves cursor to the same column on the previous line.
func MoveUp(content string, cursor int) int {
	col := GetColumnNumber(content, cursor)
	lineNum := GetLineNumber(content, cursor)

	if lineNum == 0 {
		return cursor // Already on first line
	}

	newPos := GoToLine(content, lineNum-1)
	lineEnd := LineEnd(content, newPos)
	lineLen := lineEnd - newPos

	if col > lineLen {
		return lineEnd
	}
	return newPos + col
}

// MoveDown moves cursor to the same column on the next line.
func MoveDown(content string, cursor int) int {
	col := GetColumnNumber(content, cursor)
	lineNum := GetLineNumber(content, cursor)
	totalLines := strings.Count(content, "\n")

	if lineNum >= totalLines {
		return cursor // Already on last line
	}

	newPos := GoToLine(content, lineNum+1)
	lineEnd := LineEnd(content, newPos)
	lineLen := lineEnd - newPos

	if col > lineLen {
		return lineEnd
	}
	return newPos + col
}
