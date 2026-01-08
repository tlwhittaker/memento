package ui

// EditHistory manages undo/redo history for editor content.
type EditHistory struct {
	states  []historyState // Stack of content states
	current int            // Current position in history (-1 means no history)
	maxSize int            // Maximum number of states to keep
}

type historyState struct {
	content string
	cursor  int
}

// NewEditHistory creates a new edit history with the specified max size.
func NewEditHistory(maxSize int) *EditHistory {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &EditHistory{
		states:  make([]historyState, 0, maxSize),
		current: -1,
		maxSize: maxSize,
	}
}

// Push adds a new state to the history.
// If we're not at the end of history (after undo), discard redo states.
func (h *EditHistory) Push(content string, cursor int) {
	// Don't push duplicate states
	if h.current >= 0 && h.states[h.current].content == content {
		return
	}

	// If we're not at the end, truncate redo history
	if h.current < len(h.states)-1 {
		h.states = h.states[:h.current+1]
	}

	// Add new state
	h.states = append(h.states, historyState{content: content, cursor: cursor})
	h.current = len(h.states) - 1

	// Trim if exceeding max size
	if len(h.states) > h.maxSize {
		excess := len(h.states) - h.maxSize
		h.states = h.states[excess:]
		h.current -= excess
		if h.current < 0 {
			h.current = 0
		}
	}
}

// Undo returns the previous state, if available.
func (h *EditHistory) Undo() (string, int, bool) {
	if !h.CanUndo() {
		return "", 0, false
	}

	h.current--
	state := h.states[h.current]
	return state.content, state.cursor, true
}

// Redo returns the next state, if available.
func (h *EditHistory) Redo() (string, int, bool) {
	if !h.CanRedo() {
		return "", 0, false
	}

	h.current++
	state := h.states[h.current]
	return state.content, state.cursor, true
}

// CanUndo returns true if there's a previous state to undo to.
func (h *EditHistory) CanUndo() bool {
	return h.current > 0
}

// CanRedo returns true if there's a next state to redo to.
func (h *EditHistory) CanRedo() bool {
	return h.current < len(h.states)-1
}

// Clear removes all history.
func (h *EditHistory) Clear() {
	h.states = h.states[:0]
	h.current = -1
}

// Current returns the current state without changing position.
func (h *EditHistory) Current() (string, int, bool) {
	if h.current < 0 || h.current >= len(h.states) {
		return "", 0, false
	}
	state := h.states[h.current]
	return state.content, state.cursor, true
}

// Len returns the number of states in history.
func (h *EditHistory) Len() int {
	return len(h.states)
}
