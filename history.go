package bezel

// History stores submitted lines and supports Up/Down navigation.
// Zero value is ready to use.
type History struct {
	entries []string
	idx     int
	draft   string
}

// Add appends an entry. Empty strings and consecutive duplicates are skipped.
// Resets navigation state.
func (h *History) Add(s string) {
	if s == "" {
		return
	}
	if len(h.entries) > 0 && h.entries[len(h.entries)-1] == s {
		h.Reset()
		return
	}
	h.entries = append(h.entries, s)
	h.Reset()
}

// Prev moves to the previous (older) history entry.
// On the first call after Reset, current is saved as the draft so it
// can be restored when navigating past the newest entry.
// Returns the entry text and true, or ("", false) if at the oldest.
func (h *History) Prev(current string) (string, bool) {
	if len(h.entries) == 0 || h.idx <= 0 {
		return "", false
	}
	if h.idx == len(h.entries) {
		h.draft = current
	}
	h.idx--
	return h.entries[h.idx], true
}

// Next moves to the next (newer) history entry.
// When moving past the newest entry, returns the saved draft.
// Returns the text and true, or ("", false) if not navigating.
func (h *History) Next() (string, bool) {
	if h.idx >= len(h.entries) {
		return "", false
	}
	h.idx++
	if h.idx == len(h.entries) {
		return h.draft, true
	}
	return h.entries[h.idx], true
}

// Reset stops history navigation and clears the saved draft.
func (h *History) Reset() {
	h.idx = len(h.entries)
	h.draft = ""
}

// Len returns the number of stored entries.
func (h *History) Len() int { return len(h.entries) }

// Entries returns all entries from oldest to newest.
func (h *History) Entries() []string { return h.entries }
