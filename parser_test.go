package bezel

import (
	"testing"
)

func TestParseASCII(t *testing.T) {
	p := newParser()
	events := p.Parse([]byte("abc"))

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	for i, ch := range []rune{'a', 'b', 'c'} {
		ev := events[i]
		if ev.Type != EventKey || ev.Key != KeyRune || ev.Ch != ch {
			t.Errorf("event %d: got Type=%d Key=%d Ch=%q, want EventKey/KeyRune/%q",
				i, ev.Type, ev.Key, ev.Ch, ch)
		}
	}
}

func TestParseControlChars(t *testing.T) {
	p := newParser()

	t.Run("enter", func(t *testing.T) {
		events := p.Parse([]byte{0x0d})
		if len(events) != 1 || events[0].Key != KeyEnter {
			t.Fatalf("expected Enter, got %+v", events)
		}
	})

	t.Run("tab", func(t *testing.T) {
		events := p.Parse([]byte{0x09})
		if len(events) != 1 || events[0].Key != KeyTab {
			t.Fatalf("expected Tab, got %+v", events)
		}
	})

	t.Run("backspace", func(t *testing.T) {
		events := p.Parse([]byte{0x7f})
		if len(events) != 1 || events[0].Key != KeyBackspace {
			t.Fatalf("expected Backspace, got %+v", events)
		}
	})

	t.Run("ctrl-c", func(t *testing.T) {
		events := p.Parse([]byte{0x03})
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		ev := events[0]
		if ev.Key != KeyRune || ev.Ch != 'c' || ev.Mod != ModCtrl {
			t.Errorf("got Key=%d Ch=%q Mod=%d, want KeyRune/'c'/ModCtrl", ev.Key, ev.Ch, ev.Mod)
		}
	})
}

func TestParseEscapeStandalone(t *testing.T) {
	p := newParser()
	// ESC as the only byte in the buffer = standalone Escape key.
	events := p.Parse([]byte{0x1b})
	if len(events) != 1 || events[0].Key != KeyEscape {
		t.Fatalf("expected Escape, got %+v", events)
	}
}

func TestParseArrowKeys(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		key  Key
	}{
		{"up", []byte("\033[A"), KeyUp},
		{"down", []byte("\033[B"), KeyDown},
		{"right", []byte("\033[C"), KeyRight},
		{"left", []byte("\033[D"), KeyLeft},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser()
			events := p.Parse(tt.data)
			if len(events) != 1 || events[0].Key != tt.key {
				t.Fatalf("expected %s, got %+v", tt.key, events)
			}
		})
	}
}

func TestParseModifiedKeys(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		key  Key
		mod  Modifier
	}{
		{"ctrl-right", []byte("\033[1;5C"), KeyRight, ModCtrl},
		{"ctrl-left", []byte("\033[1;5D"), KeyLeft, ModCtrl},
		{"shift-up", []byte("\033[1;2A"), KeyUp, ModShift},
		{"alt-down", []byte("\033[1;3B"), KeyDown, ModAlt},
		{"ctrl-shift-right", []byte("\033[1;6C"), KeyRight, ModCtrl | ModShift},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser()
			events := p.Parse(tt.data)
			if len(events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(events))
			}
			ev := events[0]
			if ev.Key != tt.key || ev.Mod != tt.mod {
				t.Errorf("got Key=%s Mod=%d, want Key=%s Mod=%d", ev.Key, ev.Mod, tt.key, tt.mod)
			}
		})
	}
}

func TestParseTildeKeys(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		key  Key
	}{
		{"delete", []byte("\033[3~"), KeyDelete},
		{"insert", []byte("\033[2~"), KeyInsert},
		{"page-up", []byte("\033[5~"), KeyPageUp},
		{"page-down", []byte("\033[6~"), KeyPageDown},
		{"home", []byte("\033[1~"), KeyHome},
		{"end", []byte("\033[4~"), KeyEnd},
		{"f5", []byte("\033[15~"), KeyF5},
		{"f12", []byte("\033[24~"), KeyF12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser()
			events := p.Parse(tt.data)
			if len(events) != 1 || events[0].Key != tt.key {
				t.Fatalf("expected %s, got %+v", tt.key, events)
			}
		})
	}
}

func TestParseSS3FunctionKeys(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		key  Key
	}{
		{"f1", []byte("\033OP"), KeyF1},
		{"f2", []byte("\033OQ"), KeyF2},
		{"f3", []byte("\033OR"), KeyF3},
		{"f4", []byte("\033OS"), KeyF4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser()
			events := p.Parse(tt.data)
			if len(events) != 1 || events[0].Key != tt.key {
				t.Fatalf("expected %s, got %+v", tt.key, events)
			}
		})
	}
}

func TestParseAltKeys(t *testing.T) {
	p := newParser()
	// \033 followed by 'a' in same buffer = Alt+a
	events := p.Parse([]byte{0x1b, 'a'})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	ev := events[0]
	if ev.Key != KeyRune || ev.Ch != 'a' || ev.Mod != ModAlt {
		t.Errorf("got Key=%s Ch=%q Mod=%d, want KeyRune/'a'/ModAlt", ev.Key, ev.Ch, ev.Mod)
	}
}

func TestParseUTF8(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		ch   rune
	}{
		{"e-acute", []byte{0xc3, 0xa9}, 'é'},
		{"snowman", []byte{0xe2, 0x98, 0x83}, '☃'},
		{"emoji", []byte{0xf0, 0x9f, 0x98, 0x80}, '😀'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser()
			events := p.Parse(tt.data)
			if len(events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(events))
			}
			ev := events[0]
			if ev.Key != KeyRune || ev.Ch != tt.ch {
				t.Errorf("got Key=%s Ch=%q, want KeyRune/%q", ev.Key, ev.Ch, tt.ch)
			}
		})
	}
}

func TestParseBracketedPaste(t *testing.T) {
	p := newParser()

	// Paste start + content + paste end in one buffer.
	data := []byte("\033[200~hello world\033[201~")
	events := p.Parse(data)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d: %+v", len(events), events)
	}
	ev := events[0]
	if ev.Type != EventPaste || ev.Text != "hello world" {
		t.Errorf("got Type=%d Text=%q, want EventPaste/\"hello world\"", ev.Type, ev.Text)
	}
}

func TestParseBracketedPasteAcrossBuffers(t *testing.T) {
	p := newParser()

	// Buffer 1: paste start + partial content.
	events := p.Parse([]byte("\033[200~hello"))
	if len(events) != 0 {
		t.Fatalf("expected 0 events during paste, got %d", len(events))
	}

	// Buffer 2: rest of content + paste end.
	events = p.Parse([]byte(" world\033[201~"))
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	ev := events[0]
	if ev.Type != EventPaste || ev.Text != "hello world" {
		t.Errorf("got Type=%d Text=%q, want EventPaste/\"hello world\"", ev.Type, ev.Text)
	}
}

func TestParseMixedInput(t *testing.T) {
	p := newParser()

	// "a" + Up + "b" in one buffer.
	data := []byte{'a', 0x1b, '[', 'A', 'b'}
	events := p.Parse(data)

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	if events[0].Ch != 'a' {
		t.Errorf("event 0: got Ch=%q, want 'a'", events[0].Ch)
	}
	if events[1].Key != KeyUp {
		t.Errorf("event 1: got Key=%s, want Up", events[1].Key)
	}
	if events[2].Ch != 'b' {
		t.Errorf("event 2: got Ch=%q, want 'b'", events[2].Ch)
	}
}

func TestDecodeModifier(t *testing.T) {
	tests := []struct {
		n   int
		mod Modifier
	}{
		{1, 0},                           // no modifier
		{2, ModShift},                    // 1+1
		{3, ModAlt},                      // 1+2
		{5, ModCtrl},                     // 1+4
		{6, ModCtrl | ModShift},          // 1+4+1
		{7, ModCtrl | ModAlt},            // 1+4+2
		{8, ModCtrl | ModAlt | ModShift}, // 1+4+2+1
	}

	for _, tt := range tests {
		got := decodeModifier(tt.n)
		if got != tt.mod {
			t.Errorf("decodeModifier(%d) = %d, want %d", tt.n, got, tt.mod)
		}
	}
}
