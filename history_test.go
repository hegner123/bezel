package bezel

import "testing"

func TestHistoryAddAndPrev(t *testing.T) {
	var h History
	h.Add("first")
	h.Add("second")
	h.Add("third")

	if h.Len() != 3 {
		t.Fatalf("len=%d, want 3", h.Len())
	}

	s, ok := h.Prev("current")
	if !ok || s != "third" {
		t.Fatalf("Prev1: got %q/%v, want third/true", s, ok)
	}
	s, ok = h.Prev("")
	if !ok || s != "second" {
		t.Fatalf("Prev2: got %q/%v, want second/true", s, ok)
	}
	s, ok = h.Prev("")
	if !ok || s != "first" {
		t.Fatalf("Prev3: got %q/%v, want first/true", s, ok)
	}
	// Past oldest.
	_, ok = h.Prev("")
	if ok {
		t.Fatal("Prev past oldest should return false")
	}
}

func TestHistoryNext(t *testing.T) {
	var h History
	h.Add("first")
	h.Add("second")

	// Navigate to oldest.
	h.Prev("draft")
	h.Prev("")

	s, ok := h.Next()
	if !ok || s != "second" {
		t.Fatalf("Next1: got %q/%v, want second/true", s, ok)
	}
	// Past newest — returns draft.
	s, ok = h.Next()
	if !ok || s != "draft" {
		t.Fatalf("Next2: got %q/%v, want draft/true", s, ok)
	}
	// Already at draft.
	_, ok = h.Next()
	if ok {
		t.Fatal("Next past draft should return false")
	}
}

func TestHistoryDraftPreservation(t *testing.T) {
	var h History
	h.Add("old")

	// Start typing "new stuff", then press Up.
	s, ok := h.Prev("new stuff")
	if !ok || s != "old" {
		t.Fatalf("got %q/%v, want old/true", s, ok)
	}

	// Press Down — draft restored.
	s, ok = h.Next()
	if !ok || s != "new stuff" {
		t.Fatalf("got %q/%v, want 'new stuff'/true", s, ok)
	}
}

func TestHistoryDuplicateSuppression(t *testing.T) {
	var h History
	h.Add("same")
	h.Add("same")
	h.Add("same")

	if h.Len() != 1 {
		t.Fatalf("len=%d, want 1 (duplicates suppressed)", h.Len())
	}
}

func TestHistoryEmptyStringRejected(t *testing.T) {
	var h History
	h.Add("")
	h.Add("")

	if h.Len() != 0 {
		t.Fatalf("len=%d, want 0", h.Len())
	}
}

func TestHistoryReset(t *testing.T) {
	var h History
	h.Add("one")
	h.Add("two")

	h.Prev("draft")
	h.Reset()

	// After reset, Next returns false (not navigating).
	_, ok := h.Next()
	if ok {
		t.Fatal("Next after Reset should return false")
	}

	// Prev starts fresh with a new draft.
	s, ok := h.Prev("new draft")
	if !ok || s != "two" {
		t.Fatalf("got %q/%v, want two/true", s, ok)
	}
	s, ok = h.Next()
	if !ok || s != "new draft" {
		t.Fatalf("got %q/%v, want 'new draft'/true", s, ok)
	}
}

func TestHistoryEmptyNavigation(t *testing.T) {
	var h History

	_, ok := h.Prev("anything")
	if ok {
		t.Fatal("Prev on empty history should return false")
	}
	_, ok = h.Next()
	if ok {
		t.Fatal("Next on empty history should return false")
	}
}

func TestHistoryAddResetsNavigation(t *testing.T) {
	var h History
	h.Add("old")
	h.Prev("typing")

	// Submit while navigating.
	h.Add("submitted")

	// Navigation reset — Prev gives "submitted" (newest).
	s, ok := h.Prev("fresh")
	if !ok || s != "submitted" {
		t.Fatalf("got %q/%v, want submitted/true", s, ok)
	}
}

func TestHistoryNonConsecutiveDuplicatesKept(t *testing.T) {
	var h History
	h.Add("a")
	h.Add("b")
	h.Add("a")

	if h.Len() != 3 {
		t.Fatalf("len=%d, want 3 (non-consecutive dupes kept)", h.Len())
	}
}
