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
