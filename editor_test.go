package bezel

import "testing"

func TestInsertAndString(t *testing.T) {
	var e LineEditor
	e.Insert('h')
	e.Insert('i')
	if got := e.String(); got != "hi" {
		t.Fatalf("got %q, want %q", got, "hi")
	}
	if e.Pos() != 2 || e.Len() != 2 {
		t.Fatalf("pos=%d len=%d, want 2,2", e.Pos(), e.Len())
	}
}

func TestInsertMiddle(t *testing.T) {
	var e LineEditor
	e.Insert('a')
	e.Insert('c')
	e.Left()
	e.Insert('b')
	if got := e.String(); got != "abc" {
		t.Fatalf("got %q, want %q", got, "abc")
	}
	if e.Pos() != 2 {
		t.Fatalf("pos=%d, want 2", e.Pos())
	}
}

func TestInsertString(t *testing.T) {
	var e LineEditor
	e.Insert('>')
	e.InsertString("hello")
	if got := e.String(); got != ">hello" {
		t.Fatalf("got %q, want %q", got, ">hello")
	}
	if e.Pos() != 6 {
		t.Fatalf("pos=%d, want 6", e.Pos())
	}
}

func TestBackspace(t *testing.T) {
	var e LineEditor
	e.InsertString("abc")

	t.Run("at end", func(t *testing.T) {
		if !e.Backspace() {
			t.Fatal("expected true")
		}
		if got := e.String(); got != "ab" {
			t.Fatalf("got %q, want %q", got, "ab")
		}
	})

	t.Run("at start", func(t *testing.T) {
		e.Home()
		if e.Backspace() {
			t.Fatal("expected false at start")
		}
	})
}

func TestBackspaceMiddle(t *testing.T) {
	var e LineEditor
	e.InsertString("abcd")
	e.Left()
	e.Backspace()
	if got := e.String(); got != "abd" {
		t.Fatalf("got %q, want %q", got, "abd")
	}
	if e.Pos() != 2 {
		t.Fatalf("pos=%d, want 2", e.Pos())
	}
}

func TestDelete(t *testing.T) {
	var e LineEditor
	e.InsertString("abc")
	e.Home()

	if !e.Delete() {
		t.Fatal("expected true")
	}
	if got := e.String(); got != "bc" {
		t.Fatalf("got %q, want %q", got, "bc")
	}

	e.End()
	if e.Delete() {
		t.Fatal("expected false at end")
	}
}

func TestLeftRight(t *testing.T) {
	var e LineEditor
	e.InsertString("abc")

	if e.Right() {
		t.Fatal("Right at end should return false")
	}
	e.Left()
	e.Left()
	if e.Pos() != 1 {
		t.Fatalf("pos=%d, want 1", e.Pos())
	}
	if !e.Left() {
		t.Fatal("Left should return true")
	}
	if e.Left() {
		t.Fatal("Left at start should return false")
	}
}

func TestHomeEnd(t *testing.T) {
	var e LineEditor
	e.InsertString("hello")
	e.Home()
	if e.Pos() != 0 {
		t.Fatalf("Home: pos=%d, want 0", e.Pos())
	}
	e.End()
	if e.Pos() != 5 {
		t.Fatalf("End: pos=%d, want 5", e.Pos())
	}
}

func TestWordLeft(t *testing.T) {
	var e LineEditor
	e.InsertString("hello world foo")

	e.WordLeft()
	if e.Pos() != 12 {
		t.Fatalf("pos=%d, want 12", e.Pos())
	}
	e.WordLeft()
	if e.Pos() != 6 {
		t.Fatalf("pos=%d, want 6", e.Pos())
	}
	e.WordLeft()
	if e.Pos() != 0 {
		t.Fatalf("pos=%d, want 0", e.Pos())
	}
	e.WordLeft()
	if e.Pos() != 0 {
		t.Fatalf("pos=%d, want 0 (already at start)", e.Pos())
	}
}

func TestWordRight(t *testing.T) {
	var e LineEditor
	e.InsertString("hello world foo")
	e.Home()

	e.WordRight()
	if e.Pos() != 6 {
		t.Fatalf("pos=%d, want 6", e.Pos())
	}
	e.WordRight()
	if e.Pos() != 12 {
		t.Fatalf("pos=%d, want 12", e.Pos())
	}
	e.WordRight()
	if e.Pos() != 15 {
		t.Fatalf("pos=%d, want 15", e.Pos())
	}
}

func TestDeleteToStart(t *testing.T) {
	var e LineEditor
	e.InsertString("hello world")
	e.Left()
	e.Left()
	e.Left()
	e.Left()
	e.Left()

	cut := e.DeleteToStart()
	if cut != "hello " {
		t.Fatalf("cut=%q, want %q", cut, "hello ")
	}
	if got := e.String(); got != "world" {
		t.Fatalf("got %q, want %q", got, "world")
	}
	if e.Pos() != 0 {
		t.Fatalf("pos=%d, want 0", e.Pos())
	}
}

func TestDeleteToEnd(t *testing.T) {
	var e LineEditor
	e.InsertString("hello world")
	e.Home()
	e.WordRight()

	cut := e.DeleteToEnd()
	if cut != "world" {
		t.Fatalf("cut=%q, want %q", cut, "world")
	}
	if got := e.String(); got != "hello " {
		t.Fatalf("got %q, want %q", got, "hello ")
	}
}

func TestDeleteWordBack(t *testing.T) {
	var e LineEditor
	e.InsertString("hello beautiful world")

	cut := e.DeleteWordBack()
	if cut != "world" {
		t.Fatalf("cut=%q, want %q", cut, "world")
	}
	if got := e.String(); got != "hello beautiful " {
		t.Fatalf("got %q, want %q", got, "hello beautiful ")
	}
}

func TestSubmit(t *testing.T) {
	var e LineEditor
	e.InsertString("command")
	s := e.Submit()
	if s != "command" {
		t.Fatalf("got %q, want %q", s, "command")
	}
	if !e.Empty() {
		t.Fatal("expected empty after submit")
	}
	if e.Pos() != 0 {
		t.Fatalf("pos=%d, want 0", e.Pos())
	}
}

func TestSet(t *testing.T) {
	var e LineEditor
	e.Set("preset")
	if got := e.String(); got != "preset" {
		t.Fatalf("got %q, want %q", got, "preset")
	}
	if e.Pos() != 6 {
		t.Fatalf("pos=%d, want 6", e.Pos())
	}
}

func TestClear(t *testing.T) {
	var e LineEditor
	e.InsertString("stuff")
	e.Clear()
	if !e.Empty() {
		t.Fatal("expected empty after clear")
	}
}

func TestEmptyEdgeCase(t *testing.T) {
	var e LineEditor
	if e.Backspace() {
		t.Fatal("backspace on empty")
	}
	if e.Delete() {
		t.Fatal("delete on empty")
	}
	if e.Left() {
		t.Fatal("left on empty")
	}
	if e.Right() {
		t.Fatal("right on empty")
	}
	e.WordLeft()
	e.WordRight()
	e.Home()
	e.End()
	if e.DeleteToStart() != "" {
		t.Fatal("delete to start on empty")
	}
	if e.DeleteToEnd() != "" {
		t.Fatal("delete to end on empty")
	}
	if e.DeleteWordBack() != "" {
		t.Fatal("delete word back on empty")
	}
	if e.Submit() != "" {
		t.Fatal("submit on empty")
	}
}

func TestInsertNewline(t *testing.T) {
	var e LineEditor
	e.InsertString("hello")
	e.InsertNewline()
	e.InsertString("world")
	if got := e.String(); got != "hello\nworld" {
		t.Fatalf("got %q, want %q", got, "hello\nworld")
	}
	if e.Len() != 11 {
		t.Fatalf("len=%d, want 11", e.Len())
	}
}

func TestLines(t *testing.T) {
	var e LineEditor
	e.InsertString("a\nb\nc")
	lines := e.Lines()
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	want := []string{"a", "b", "c"}
	for i, w := range want {
		if lines[i] != w {
			t.Errorf("line %d: got %q, want %q", i, lines[i], w)
		}
	}
}

func TestRowCol(t *testing.T) {
	var e LineEditor
	e.InsertString("ab\ncde\nf")

	if r, c := e.CursorPos(); r != 2 || c != 1 {
		t.Fatalf("at end: row=%d col=%d, want 2,1", r, c)
	}

	e.Home()
	if r, c := e.CursorPos(); r != 2 || c != 0 {
		t.Fatalf("home on line 3: row=%d col=%d, want 2,0", r, c)
	}

	e.Up()
	e.End()
	if r, c := e.CursorPos(); r != 1 || c != 3 {
		t.Fatalf("end of line 2: row=%d col=%d, want 1,3", r, c)
	}
}

func TestUpDown(t *testing.T) {
	var e LineEditor
	e.InsertString("abc\ndef\nghi")
	// Cursor at end of "ghi", pos=11, row=2, col=3
	e.Home() // pos=8 (start of "ghi")
	e.Right()
	e.Right() // pos=10, col=2, at 'i'

	if !e.Up() {
		t.Fatal("Up from line 2 should return true")
	}
	if r, c := e.CursorPos(); r != 1 || c != 2 {
		t.Fatalf("after Up: row=%d col=%d, want 1,2", r, c)
	}

	if !e.Up() {
		t.Fatal("Up from line 1 should return true")
	}
	if r, c := e.CursorPos(); r != 0 || c != 2 {
		t.Fatalf("after 2nd Up: row=%d col=%d, want 0,2", r, c)
	}

	if e.Up() {
		t.Fatal("Up from line 0 should return false")
	}

	if !e.Down() {
		t.Fatal("Down from line 0 should return true")
	}
	if r, c := e.CursorPos(); r != 1 || c != 2 {
		t.Fatalf("after Down: row=%d col=%d, want 1,2", r, c)
	}
}

func TestUpDownClampCol(t *testing.T) {
	var e LineEditor
	e.InsertString("abcdef\nhi\njklmno")
	// line 0: "abcdef" (6), line 1: "hi" (2), line 2: "jklmno" (6)
	// Cursor at end of "jklmno", col=6

	e.Up() // to line 1, col clamped to 2
	if r, c := e.CursorPos(); r != 1 || c != 2 {
		t.Fatalf("up to short line: row=%d col=%d, want 1,2", r, c)
	}

	e.Up() // to line 0, col stays 2 (line 0 has 6 chars)
	if r, c := e.CursorPos(); r != 0 || c != 2 {
		t.Fatalf("up to long line: row=%d col=%d, want 0,2", r, c)
	}
}

func TestDownOnLastLine(t *testing.T) {
	var e LineEditor
	e.InsertString("only line")
	if e.Down() {
		t.Fatal("Down on single line should return false")
	}
}

func TestHomeEndMultiline(t *testing.T) {
	var e LineEditor
	e.InsertString("hello\nworld")
	// Cursor at end of "world"

	e.Home()
	if e.Col() != 0 || e.Row() != 1 {
		t.Fatalf("Home: row=%d col=%d, want 1,0", e.Row(), e.Col())
	}

	e.End()
	if e.Col() != 5 || e.Row() != 1 {
		t.Fatalf("End: row=%d col=%d, want 1,5", e.Row(), e.Col())
	}

	e.Up()
	e.Home()
	if e.Pos() != 0 {
		t.Fatalf("Home on first line: pos=%d, want 0", e.Pos())
	}
	e.End()
	if e.Pos() != 5 {
		t.Fatalf("End on first line: pos=%d, want 5", e.Pos())
	}
}

func TestDeleteToStartMultiline(t *testing.T) {
	var e LineEditor
	e.InsertString("hello\nworld")
	e.Left()
	e.Left()
	// pos=9, at 'l' in "world", col=3

	cut := e.DeleteToStart()
	if cut != "wor" {
		t.Fatalf("cut=%q, want %q", cut, "wor")
	}
	if got := e.String(); got != "hello\nld" {
		t.Fatalf("got %q, want %q", got, "hello\nld")
	}
}

func TestDeleteToEndMultiline(t *testing.T) {
	var e LineEditor
	e.InsertString("hello\nworld")
	e.Up()
	e.Home()
	e.Right()
	e.Right()
	// pos=2, at first 'l' in "hello", col=2

	cut := e.DeleteToEnd()
	if cut != "llo" {
		t.Fatalf("cut=%q, want %q", cut, "llo")
	}
	if got := e.String(); got != "he\nworld" {
		t.Fatalf("got %q, want %q", got, "he\nworld")
	}
}

func TestWordLeftNewlineBoundary(t *testing.T) {
	var e LineEditor
	e.InsertString("hello\nworld")
	// pos=11, at end of "world"

	e.WordLeft() // should go to start of "world"
	if e.Pos() != 6 {
		t.Fatalf("pos=%d, want 6", e.Pos())
	}

	e.WordLeft() // should cross newline to start of "hello"
	if e.Pos() != 0 {
		t.Fatalf("pos=%d, want 0", e.Pos())
	}
}

func TestWordRightNewlineBoundary(t *testing.T) {
	var e LineEditor
	e.InsertString("hello\nworld")
	e.Home()
	e.Up()
	e.Home()
	// pos=0, at start of "hello"

	e.WordRight() // should cross newline to start of "world"
	if e.Pos() != 6 {
		t.Fatalf("pos=%d, want 6", e.Pos())
	}
}

func TestBackspaceJoinsLines(t *testing.T) {
	var e LineEditor
	e.InsertString("hello\nworld")
	e.Home() // start of "world", pos=6
	e.Backspace()
	if got := e.String(); got != "helloworld" {
		t.Fatalf("got %q, want %q", got, "helloworld")
	}
	if e.Pos() != 5 {
		t.Fatalf("pos=%d, want 5", e.Pos())
	}
}

func TestDeleteWordBackMultiline(t *testing.T) {
	var e LineEditor
	e.InsertString("hello\nworld")
	e.Home() // start of "world"

	// At start of line, DeleteWordBack should cross the newline
	cut := e.DeleteWordBack()
	if cut != "\nhello" && cut != "hello\n" {
		// Should delete backwards: first skip \n (whitespace), then skip "hello"
		// From pos=6, buf[5]='\n' → skip → pos=5. buf[4]='o' → skip word → pos=0
		// cut = buf[0:6] = "hello\n"
		t.Fatalf("cut=%q, want %q", cut, "hello\n")
	}
	if got := e.String(); got != "world" {
		t.Fatalf("got %q, want %q", got, "world")
	}
}

func TestActionUpHistoryFallback(t *testing.T) {
	var e LineEditor
	var hist History
	hist.Add("previous")
	km := DefaultKeyMap()

	e.InsertString("current")

	// Up arrow on single-line editor should fall through to history
	ev := Event{Type: EventKey, Key: KeyUp}
	action, _ := e.HandleEvent(ev, km, &hist)
	if action != ActionHistoryPrev {
		t.Fatalf("got action %d, want ActionHistoryPrev(%d)", action, ActionHistoryPrev)
	}
	if got := e.String(); got != "previous" {
		t.Fatalf("got %q, want %q", got, "previous")
	}
}

func TestActionUpLineNav(t *testing.T) {
	var e LineEditor
	var hist History
	hist.Add("previous")
	km := DefaultKeyMap()

	e.InsertString("line1\nline2")

	// Up arrow on second line should navigate, not history
	ev := Event{Type: EventKey, Key: KeyUp}
	action, _ := e.HandleEvent(ev, km, &hist)
	if action != ActionUp {
		t.Fatalf("got action %d, want ActionUp(%d)", action, ActionUp)
	}
	if e.Row() != 0 {
		t.Fatalf("row=%d, want 0", e.Row())
	}
}

func TestActionNewline(t *testing.T) {
	var e LineEditor
	km := DefaultKeyMap()

	e.InsertString("hello")

	// Alt+Enter should insert newline
	ev := Event{Type: EventKey, Key: KeyEnter, Mod: ModAlt}
	action, _ := e.HandleEvent(ev, km, nil)
	if action != ActionNewline {
		t.Fatalf("got action %d, want ActionNewline(%d)", action, ActionNewline)
	}
	if got := e.String(); got != "hello\n" {
		t.Fatalf("got %q, want %q", got, "hello\n")
	}
}

func TestVisualNoWrap(t *testing.T) {
	var e LineEditor
	e.InsertString("hello")
	v := e.Visual(80, []string{"> "})

	if len(v.Rows) != 1 {
		t.Fatalf("rows=%d, want 1", len(v.Rows))
	}
	want := "> hello" + CursorBlock
	if v.Rows[0] != want {
		t.Fatalf("row 0=%q, want %q", v.Rows[0], want)
	}
}

func TestVisualWrapSingleLine(t *testing.T) {
	var e LineEditor
	// prefix "> " (2 chars), width 10, avail=8
	e.InsertString("abcdefghij") // 10 chars, exceeds avail
	v := e.Visual(10, []string{"> "})

	// Content with cursor: "abcdefghij█" (11 runes)
	// First visual row: "> abcdefgh" (prefix + 8)
	// Second visual row: "ij█" (remaining 3)
	if len(v.Rows) != 2 {
		t.Fatalf("rows=%d, want 2", len(v.Rows))
	}
	if v.Rows[0] != "> abcdefgh" {
		t.Fatalf("row 0=%q, want %q", v.Rows[0], "> abcdefgh")
	}
	want1 := "ij" + CursorBlock
	if v.Rows[1] != want1 {
		t.Fatalf("row 1=%q, want %q", v.Rows[1], want1)
	}
}

func TestVisualWrapExactBoundary(t *testing.T) {
	var e LineEditor
	// prefix "> " (2), width 10, avail=8
	e.InsertString("abcdefgh") // exactly avail chars
	// cursor at col 8 == avail, cursor block wraps to next row
	v := e.Visual(10, []string{"> "})

	if len(v.Rows) != 2 {
		t.Fatalf("rows=%d, want 2", len(v.Rows))
	}
	if v.Rows[0] != "> abcdefgh" {
		t.Fatalf("row 0=%q, want %q", v.Rows[0], "> abcdefgh")
	}
	if v.Rows[1] != CursorBlock {
		t.Fatalf("row 1=%q, want %q", v.Rows[1], CursorBlock)
	}
}

func TestVisualWrapCursorMiddle(t *testing.T) {
	var e LineEditor
	// prefix "> " (2), width 10, avail=8
	e.InsertString("abcdefghij") // 10 chars, wraps
	e.Home()
	e.Right()
	e.Right()
	e.Right()
	e.Right()
	e.Right() // pos=5, cursor block at position 5

	v := e.Visual(10, []string{"> "})
	// Content with cursor: "abcde█fghij" (11 runes)
	// First 8: "abcde█fg" → "> abcde█fg"
	// Remaining: "hij"
	want0 := "> abcde" + CursorBlock + "fg"
	if v.Rows[0] != want0 {
		t.Fatalf("row 0=%q, want %q", v.Rows[0], want0)
	}
	if v.Rows[1] != "hij" {
		t.Fatalf("row 1=%q, want %q", v.Rows[1], "hij")
	}
}

func TestVisualMultiline(t *testing.T) {
	var e LineEditor
	e.InsertString("hello\nworld")
	v := e.Visual(80, []string{"> ", "  "})

	if len(v.Rows) != 2 {
		t.Fatalf("rows=%d, want 2", len(v.Rows))
	}
	if v.Rows[0] != "> hello" {
		t.Fatalf("row 0=%q, want %q", v.Rows[0], "> hello")
	}
	want1 := "  world" + CursorBlock
	if v.Rows[1] != want1 {
		t.Fatalf("row 1=%q, want %q", v.Rows[1], want1)
	}
}

func TestVisualMultilineWrap(t *testing.T) {
	var e LineEditor
	// width=10, first prefix "> " (2), second "  " (2), avail=8 for both
	e.InsertString("short\nabcdefghij") // line 2 wraps
	v := e.Visual(10, []string{"> ", "  "})

	// line 0: "> short" (1 visual row)
	// line 1: "abcdefghij█" with prefix "  ", avail=8
	//   "  abcdefgh" + "ij█" (2 visual rows)
	if len(v.Rows) != 3 {
		t.Fatalf("rows=%d, want 3", len(v.Rows))
	}
	if v.Rows[0] != "> short" {
		t.Fatalf("row 0=%q", v.Rows[0])
	}
	if v.Rows[1] != "  abcdefgh" {
		t.Fatalf("row 1=%q", v.Rows[1])
	}
	want2 := "ij" + CursorBlock
	if v.Rows[2] != want2 {
		t.Fatalf("row 2=%q, want %q", v.Rows[2], want2)
	}
}

func TestVisualEmpty(t *testing.T) {
	var e LineEditor
	v := e.Visual(80, []string{"> "})

	if len(v.Rows) != 1 {
		t.Fatalf("rows=%d, want 1", len(v.Rows))
	}
	want := "> " + CursorBlock
	if v.Rows[0] != want {
		t.Fatalf("row 0=%q, want %q", v.Rows[0], want)
	}
}

func TestUTF8(t *testing.T) {
	var e LineEditor
	e.Insert('é')
	e.Insert('☃')
	e.Insert('😀')
	if got := e.String(); got != "é☃😀" {
		t.Fatalf("got %q, want %q", got, "é☃😀")
	}
	if e.Len() != 3 {
		t.Fatalf("len=%d, want 3", e.Len())
	}
	e.Left()
	e.Backspace()
	if got := e.String(); got != "é😀" {
		t.Fatalf("got %q, want %q", got, "é😀")
	}
}
