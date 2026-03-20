package bezel

import (
	"slices"
	"strings"
)

// CursorBlock is the pseudo cursor character rendered in the bezel.
const CursorBlock = "█"

// LineEditor manages an editable line of text with cursor position.
// Zero value is ready to use.
type LineEditor struct {
	buf []rune
	pos int
}

// lineStart returns the buffer index of the first rune on the current line.
func (e *LineEditor) lineStart() int {
	for i := e.pos - 1; i >= 0; i-- {
		if e.buf[i] == '\n' {
			return i + 1
		}
	}
	return 0
}

// lineEnd returns the buffer index just past the last rune on the current
// line (the position of the next '\n', or len(buf) if on the last line).
func (e *LineEditor) lineEnd() int {
	for i := e.pos; i < len(e.buf); i++ {
		if e.buf[i] == '\n' {
			return i
		}
	}
	return len(e.buf)
}

// Insert adds a rune at the cursor and advances the cursor.
func (e *LineEditor) Insert(r rune) {
	e.buf = slices.Insert(e.buf, e.pos, r)
	e.pos++
}

// InsertString adds all runes from s at the cursor.
func (e *LineEditor) InsertString(s string) {
	runes := []rune(s)
	e.buf = slices.Insert(e.buf, e.pos, runes...)
	e.pos += len(runes)
}

// InsertNewline inserts a newline at the cursor, starting a new line.
func (e *LineEditor) InsertNewline() {
	e.Insert('\n')
}

// Backspace deletes the rune before the cursor.
func (e *LineEditor) Backspace() bool {
	if e.pos == 0 {
		return false
	}
	e.pos--
	e.buf = slices.Delete(e.buf, e.pos, e.pos+1)
	return true
}

// Delete deletes the rune at the cursor.
func (e *LineEditor) Delete() bool {
	if e.pos >= len(e.buf) {
		return false
	}
	e.buf = slices.Delete(e.buf, e.pos, e.pos+1)
	return true
}

// Left moves the cursor one position left.
func (e *LineEditor) Left() bool {
	if e.pos == 0 {
		return false
	}
	e.pos--
	return true
}

// Right moves the cursor one position right.
func (e *LineEditor) Right() bool {
	if e.pos >= len(e.buf) {
		return false
	}
	e.pos++
	return true
}

// Up moves the cursor to the same column on the previous line.
// Returns false if already on the first line.
func (e *LineEditor) Up() bool {
	ls := e.lineStart()
	if ls == 0 {
		return false
	}
	col := e.pos - ls
	prevStart := 0
	for i := ls - 2; i >= 0; i-- {
		if e.buf[i] == '\n' {
			prevStart = i + 1
			break
		}
	}
	prevLen := (ls - 1) - prevStart
	if col > prevLen {
		col = prevLen
	}
	e.pos = prevStart + col
	return true
}

// Down moves the cursor to the same column on the next line.
// Returns false if already on the last line.
func (e *LineEditor) Down() bool {
	le := e.lineEnd()
	if le >= len(e.buf) {
		return false
	}
	col := e.pos - e.lineStart()
	nextStart := le + 1
	nextEnd := len(e.buf)
	for i := nextStart; i < len(e.buf); i++ {
		if e.buf[i] == '\n' {
			nextEnd = i
			break
		}
	}
	nextLen := nextEnd - nextStart
	if col > nextLen {
		col = nextLen
	}
	e.pos = nextStart + col
	return true
}

// Home moves the cursor to the start of the current line.
func (e *LineEditor) Home() { e.pos = e.lineStart() }

// End moves the cursor to the end of the current line.
func (e *LineEditor) End() { e.pos = e.lineEnd() }

// WordLeft moves the cursor to the start of the previous word.
// Newlines are treated as whitespace boundaries.
func (e *LineEditor) WordLeft() {
	for e.pos > 0 && (e.buf[e.pos-1] == ' ' || e.buf[e.pos-1] == '\n') {
		e.pos--
	}
	for e.pos > 0 && e.buf[e.pos-1] != ' ' && e.buf[e.pos-1] != '\n' {
		e.pos--
	}
}

// WordRight moves the cursor past the end of the next word.
// Newlines are treated as whitespace boundaries.
func (e *LineEditor) WordRight() {
	n := len(e.buf)
	for e.pos < n && e.buf[e.pos] != ' ' && e.buf[e.pos] != '\n' {
		e.pos++
	}
	for e.pos < n && (e.buf[e.pos] == ' ' || e.buf[e.pos] == '\n') {
		e.pos++
	}
}

// DeleteToStart removes everything from the cursor to the start of the
// current line. Returns the deleted text.
func (e *LineEditor) DeleteToStart() string {
	ls := e.lineStart()
	if e.pos == ls {
		return ""
	}
	cut := string(e.buf[ls:e.pos])
	e.buf = slices.Delete(e.buf, ls, e.pos)
	e.pos = ls
	return cut
}

// DeleteToEnd removes everything from the cursor to the end of the
// current line. Returns the deleted text.
func (e *LineEditor) DeleteToEnd() string {
	le := e.lineEnd()
	if e.pos >= le {
		return ""
	}
	cut := string(e.buf[e.pos:le])
	e.buf = slices.Delete(e.buf, e.pos, le)
	return cut
}

// DeleteWordBack removes the word before the cursor.
// Newlines are treated as whitespace boundaries.
// Returns the deleted text.
func (e *LineEditor) DeleteWordBack() string {
	if e.pos == 0 {
		return ""
	}
	start := e.pos
	for e.pos > 0 && (e.buf[e.pos-1] == ' ' || e.buf[e.pos-1] == '\n') {
		e.pos--
	}
	for e.pos > 0 && e.buf[e.pos-1] != ' ' && e.buf[e.pos-1] != '\n' {
		e.pos--
	}
	cut := string(e.buf[e.pos:start])
	e.buf = slices.Delete(e.buf, e.pos, start)
	return cut
}

// Submit returns the current content and resets the editor.
func (e *LineEditor) Submit() string {
	s := string(e.buf)
	e.buf = e.buf[:0]
	e.pos = 0
	return s
}

// Set replaces the content and moves the cursor to the end.
func (e *LineEditor) Set(s string) {
	e.buf = []rune(s)
	e.pos = len(e.buf)
}

// Clear empties the editor.
func (e *LineEditor) Clear() {
	e.buf = e.buf[:0]
	e.pos = 0
}

// String returns the current content.
func (e *LineEditor) String() string { return string(e.buf) }

// Pos returns the cursor position (in runes from start).
func (e *LineEditor) Pos() int { return e.pos }

// Len returns the content length in runes.
func (e *LineEditor) Len() int { return len(e.buf) }

// Empty reports whether the editor has no content.
func (e *LineEditor) Empty() bool { return len(e.buf) == 0 }

// StringWithCursor returns the content with CursorBlock (█) inserted at
// the cursor position. Use this when building bezel lines for Redraw.
func (e *LineEditor) StringWithCursor() string {
	left := string(e.buf[:e.pos])
	right := string(e.buf[e.pos:])
	return left + CursorBlock + right
}

// Row returns the 0-indexed line number of the cursor.
func (e *LineEditor) Row() int {
	row := 0
	for i := range e.pos {
		if e.buf[i] == '\n' {
			row++
		}
	}
	return row
}

// Col returns the 0-indexed column of the cursor within the current line.
func (e *LineEditor) Col() int { return e.pos - e.lineStart() }

// CursorPos returns the cursor position as (row, col), both 0-indexed.
func (e *LineEditor) CursorPos() (row, col int) { return e.Row(), e.Col() }

// Lines returns the editor content split into lines.
func (e *LineEditor) Lines() []string { return strings.Split(string(e.buf), "\n") }

// VisualInfo describes the editor content laid out for rendering in a
// terminal of a given width, with CursorBlock embedded at the cursor
// position.
type VisualInfo struct {
	Rows []string // Visual rows (wrapped, with prefixes and cursor block applied).
}

// Visual computes visual rows accounting for line wrapping within the
// given terminal width. CursorBlock (█) is embedded at the cursor
// position. prefixes[i] is prepended to logical line i (e.g. a prompt
// or continuation marker); missing entries default to "".
// The caller passes Rows directly to Bezel.Redraw.
func (e *LineEditor) Visual(width int, prefixes []string) VisualInfo {
	if width <= 0 {
		width = 80
	}

	// Build content with cursor block inserted.
	content := make([]rune, 0, len(e.buf)+1)
	content = append(content, e.buf[:e.pos]...)
	content = append(content, []rune(CursorBlock)...)
	content = append(content, e.buf[e.pos:]...)

	lines := strings.Split(string(content), "\n")

	var info VisualInfo

	for i, line := range lines {
		prefix := ""
		if i < len(prefixes) {
			prefix = prefixes[i]
		}

		plen := len([]rune(prefix))
		runes := []rune(line)
		avail := width - plen
		if avail < 1 {
			avail = 1
		}

		if len(runes) <= avail {
			info.Rows = append(info.Rows, prefix+string(runes))
		} else {
			info.Rows = append(info.Rows, prefix+string(runes[:avail]))
			for start := avail; start < len(runes); start += width {
				end := start + width
				if end > len(runes) {
					end = len(runes)
				}
				info.Rows = append(info.Rows, string(runes[start:end]))
			}
		}
	}

	return info
}

// HandleEvent processes an input event using the given keymap and
// optional history. Pass nil for hist if history is not needed.
// Returns the action taken and, for ActionSubmit/ActionPaste, the text.
// Unmodified printable runes are inserted without a keymap entry.
// Paste events are always handled.
func (e *LineEditor) HandleEvent(ev Event, km KeyMap, hist *History) (Action, string) {
	if ev.Type == EventPaste {
		e.InsertString(ev.Text)
		return ActionPaste, ev.Text
	}
	if ev.Type != EventKey {
		return ActionNone, ""
	}

	bind := KeyBind{Key: ev.Key, Mod: ev.Mod}
	if ev.Key == KeyRune {
		bind.Ch = ev.Ch
	}

	if action, ok := km[bind]; ok {
		return e.execAction(action, hist)
	}

	if ev.Key == KeyRune && ev.Mod == 0 {
		e.Insert(ev.Ch)
		return ActionInsert, ""
	}

	return ActionNone, ""
}

func (e *LineEditor) execAction(a Action, hist *History) (Action, string) {
	switch a {
	case ActionBackspace:
		e.Backspace()
	case ActionDelete:
		e.Delete()
	case ActionLeft:
		e.Left()
	case ActionRight:
		e.Right()
	case ActionWordLeft:
		e.WordLeft()
	case ActionWordRight:
		e.WordRight()
	case ActionHome:
		e.Home()
	case ActionEnd:
		e.End()
	case ActionDeleteToStart:
		e.DeleteToStart()
	case ActionDeleteToEnd:
		e.DeleteToEnd()
	case ActionDeleteWordBack:
		e.DeleteWordBack()
	case ActionSubmit:
		return ActionSubmit, e.Submit()
	case ActionNewline:
		e.InsertNewline()
	case ActionUp:
		if e.Up() {
			return ActionUp, ""
		}
		if hist == nil {
			return ActionNone, ""
		}
		if s, ok := hist.Prev(e.String()); ok {
			e.Set(s)
			return ActionHistoryPrev, ""
		}
		return ActionNone, ""
	case ActionDown:
		if e.Down() {
			return ActionDown, ""
		}
		if hist == nil {
			return ActionNone, ""
		}
		if s, ok := hist.Next(); ok {
			e.Set(s)
			return ActionHistoryNext, ""
		}
		return ActionNone, ""
	case ActionHistoryPrev:
		if hist == nil {
			return ActionNone, ""
		}
		if s, ok := hist.Prev(e.String()); ok {
			e.Set(s)
			return ActionHistoryPrev, ""
		}
		return ActionNone, ""
	case ActionHistoryNext:
		if hist == nil {
			return ActionNone, ""
		}
		if s, ok := hist.Next(); ok {
			e.Set(s)
			return ActionHistoryNext, ""
		}
		return ActionNone, ""
	}
	return a, ""
}
