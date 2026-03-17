package bezel

import "slices"

// LineEditor manages an editable line of text with cursor position.
// Zero value is ready to use.
type LineEditor struct {
	buf []rune
	pos int
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

// Home moves the cursor to the start.
func (e *LineEditor) Home() { e.pos = 0 }

// End moves the cursor to the end.
func (e *LineEditor) End() { e.pos = len(e.buf) }

// WordLeft moves the cursor to the start of the previous word.
func (e *LineEditor) WordLeft() {
	for e.pos > 0 && e.buf[e.pos-1] == ' ' {
		e.pos--
	}
	for e.pos > 0 && e.buf[e.pos-1] != ' ' {
		e.pos--
	}
}

// WordRight moves the cursor past the end of the next word.
func (e *LineEditor) WordRight() {
	n := len(e.buf)
	for e.pos < n && e.buf[e.pos] != ' ' {
		e.pos++
	}
	for e.pos < n && e.buf[e.pos] == ' ' {
		e.pos++
	}
}

// DeleteToStart removes everything from the cursor to the start of the line.
// Returns the deleted text.
func (e *LineEditor) DeleteToStart() string {
	if e.pos == 0 {
		return ""
	}
	cut := string(e.buf[:e.pos])
	e.buf = slices.Delete(e.buf, 0, e.pos)
	e.pos = 0
	return cut
}

// DeleteToEnd removes everything from the cursor to the end of the line.
// Returns the deleted text.
func (e *LineEditor) DeleteToEnd() string {
	if e.pos >= len(e.buf) {
		return ""
	}
	cut := string(e.buf[e.pos:])
	e.buf = e.buf[:e.pos]
	return cut
}

// DeleteWordBack removes the word before the cursor.
// Returns the deleted text.
func (e *LineEditor) DeleteWordBack() string {
	if e.pos == 0 {
		return ""
	}
	start := e.pos
	for e.pos > 0 && e.buf[e.pos-1] == ' ' {
		e.pos--
	}
	for e.pos > 0 && e.buf[e.pos-1] != ' ' {
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
