package bezel

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

type parser struct {
	inPaste  bool
	pasteBuf []byte
	pasteRaw []byte
}

func newParser() *parser {
	return &parser{}
}

// Parse processes a batch of raw bytes and returns structured events.
// Escape sequences are expected to arrive complete within a single read
// (standard behavior for local terminals).
func (p *parser) Parse(data []byte) []Event {
	var events []Event
	i := 0
	for i < len(data) {
		if p.inPaste {
			n, ev := p.parsePaste(data[i:])
			if ev != nil {
				events = append(events, *ev)
			}
			i += n
			continue
		}

		b := data[i]
		switch {
		case b == 0x1b:
			ev, n := p.parseEscape(data[i:])
			if ev != nil {
				events = append(events, *ev)
			}
			i += n
		case b == 0x09:
			events = append(events, Event{Type: EventKey, Key: KeyTab, Raw: []byte{b}})
			i++
		case b == 0x0d:
			events = append(events, Event{Type: EventKey, Key: KeyEnter, Raw: []byte{b}})
			i++
		case b == 0x7f:
			events = append(events, Event{Type: EventKey, Key: KeyBackspace, Raw: []byte{b}})
			i++
		case b < 0x20:
			events = append(events, Event{
				Type: EventKey, Key: KeyRune,
				Ch: controlRune(b), Mod: ModCtrl,
				Raw: []byte{b},
			})
			i++
		case b < 0x80:
			events = append(events, Event{
				Type: EventKey, Key: KeyRune,
				Ch: rune(b), Raw: []byte{b},
			})
			i++
		default:
			ev, n := p.parseUTF8(data[i:])
			events = append(events, ev)
			i += n
		}
	}
	return events
}

// controlRune maps a control byte (0x01-0x1A) to its letter.
func controlRune(b byte) rune {
	if b == 0 {
		return '@'
	}
	return rune('a' + b - 1)
}

func (p *parser) parseEscape(data []byte) (*Event, int) {
	if len(data) == 1 {
		ev := Event{Type: EventKey, Key: KeyEscape, Raw: []byte{0x1b}}
		return &ev, 1
	}

	switch data[1] {
	case '[':
		return p.parseCSI(data)
	case 'O':
		return p.parseSS3(data)
	default:
		return p.parseAltKey(data)
	}
}

func (p *parser) parseCSI(data []byte) (*Event, int) {
	// data[0]=0x1b, data[1]='['
	i := 2

	// Collect parameter bytes (digits and semicolons).
	paramStart := i
	for i < len(data) && ((data[i] >= '0' && data[i] <= '9') || data[i] == ';') {
		i++
	}
	params := string(data[paramStart:i])

	// Skip intermediate bytes (0x20-0x2F), rare in keyboard input.
	for i < len(data) && data[i] >= 0x20 && data[i] <= 0x2F {
		i++
	}

	if i >= len(data) {
		// Incomplete sequence in this buffer — emit as raw.
		raw := make([]byte, len(data))
		copy(raw, data)
		ev := Event{Raw: raw}
		return &ev, len(data)
	}

	final := data[i]
	i++

	raw := make([]byte, i)
	copy(raw, data[:i])

	return p.mapCSI(params, final, raw), i
}

func (p *parser) mapCSI(params string, final byte, raw []byte) *Event {
	pp := splitParams(params)

	// Tilde sequences: \033[N~ or \033[N;mod~
	if final == '~' && len(pp) > 0 {
		if pp[0] == 200 {
			p.inPaste = true
			p.pasteBuf = p.pasteBuf[:0]
			p.pasteRaw = append(p.pasteRaw[:0], raw...)
			return nil
		}
		if pp[0] == 201 {
			return nil
		}

		var mod Modifier
		if len(pp) >= 2 {
			mod = decodeModifier(pp[1])
		}

		key := tildeKey(pp[0])
		if key != 0 {
			return &Event{Type: EventKey, Key: key, Mod: mod, Raw: raw}
		}
		return &Event{Raw: raw}
	}

	// Letter-final sequences: \033[A or \033[1;5A
	var mod Modifier
	if len(pp) >= 2 {
		mod = decodeModifier(pp[1])
	}

	switch final {
	case 'A':
		return &Event{Type: EventKey, Key: KeyUp, Mod: mod, Raw: raw}
	case 'B':
		return &Event{Type: EventKey, Key: KeyDown, Mod: mod, Raw: raw}
	case 'C':
		return &Event{Type: EventKey, Key: KeyRight, Mod: mod, Raw: raw}
	case 'D':
		return &Event{Type: EventKey, Key: KeyLeft, Mod: mod, Raw: raw}
	case 'H':
		return &Event{Type: EventKey, Key: KeyHome, Mod: mod, Raw: raw}
	case 'F':
		return &Event{Type: EventKey, Key: KeyEnd, Mod: mod, Raw: raw}
	default:
		return &Event{Raw: raw}
	}
}

func (p *parser) parseSS3(data []byte) (*Event, int) {
	// data[0]=0x1b, data[1]='O'
	if len(data) < 3 {
		raw := make([]byte, len(data))
		copy(raw, data)
		ev := Event{Raw: raw}
		return &ev, len(data)
	}

	raw := []byte{data[0], data[1], data[2]}

	switch data[2] {
	case 'P':
		return &Event{Type: EventKey, Key: KeyF1, Raw: raw}, 3
	case 'Q':
		return &Event{Type: EventKey, Key: KeyF2, Raw: raw}, 3
	case 'R':
		return &Event{Type: EventKey, Key: KeyF3, Raw: raw}, 3
	case 'S':
		return &Event{Type: EventKey, Key: KeyF4, Raw: raw}, 3
	case 'A':
		return &Event{Type: EventKey, Key: KeyUp, Raw: raw}, 3
	case 'B':
		return &Event{Type: EventKey, Key: KeyDown, Raw: raw}, 3
	case 'C':
		return &Event{Type: EventKey, Key: KeyRight, Raw: raw}, 3
	case 'D':
		return &Event{Type: EventKey, Key: KeyLeft, Raw: raw}, 3
	case 'H':
		return &Event{Type: EventKey, Key: KeyHome, Raw: raw}, 3
	case 'F':
		return &Event{Type: EventKey, Key: KeyEnd, Raw: raw}, 3
	default:
		return &Event{Raw: raw}, 3
	}
}

func (p *parser) parseAltKey(data []byte) (*Event, int) {
	b := data[1]
	raw := []byte{0x1b, b}

	switch {
	case b == 0x7f:
		return &Event{Type: EventKey, Key: KeyBackspace, Mod: ModAlt, Raw: raw}, 2
	case b == 0x0d:
		return &Event{Type: EventKey, Key: KeyEnter, Mod: ModAlt, Raw: raw}, 2
	case b < 0x20:
		return &Event{
			Type: EventKey, Key: KeyRune,
			Ch: controlRune(b), Mod: ModAlt | ModCtrl,
			Raw: raw,
		}, 2
	case b < 0x80:
		return &Event{
			Type: EventKey, Key: KeyRune,
			Ch: rune(b), Mod: ModAlt,
			Raw: raw,
		}, 2
	default:
		// High byte after ESC — emit ESC alone, let UTF-8 handle the byte.
		return &Event{Type: EventKey, Key: KeyEscape, Raw: []byte{0x1b}}, 1
	}
}

func (p *parser) parseUTF8(data []byte) (Event, int) {
	r, size := utf8.DecodeRune(data)
	if r == utf8.RuneError && size <= 1 {
		return Event{Raw: []byte{data[0]}}, 1
	}
	raw := make([]byte, size)
	copy(raw, data[:size])
	return Event{Type: EventKey, Key: KeyRune, Ch: r, Raw: raw}, size
}

// parsePaste accumulates bytes until the bracketed paste end marker
// \033[201~ is found. Returns bytes consumed and an event (nil if
// the end marker hasn't arrived yet).
func (p *parser) parsePaste(data []byte) (int, *Event) {
	for i, b := range data {
		p.pasteBuf = append(p.pasteBuf, b)
		p.pasteRaw = append(p.pasteRaw, b)

		if b == '~' && len(p.pasteBuf) >= 6 {
			end := len(p.pasteBuf)
			if p.pasteBuf[end-6] == 0x1b &&
				p.pasteBuf[end-5] == '[' &&
				p.pasteBuf[end-4] == '2' &&
				p.pasteBuf[end-3] == '0' &&
				p.pasteBuf[end-2] == '1' &&
				p.pasteBuf[end-1] == '~' {

				content := string(p.pasteBuf[:end-6])
				raw := make([]byte, len(p.pasteRaw))
				copy(raw, p.pasteRaw)
				ev := Event{
					Type: EventPaste,
					Text: content,
					Raw:  raw,
				}
				p.inPaste = false
				p.pasteBuf = p.pasteBuf[:0]
				p.pasteRaw = p.pasteRaw[:0]
				return i + 1, &ev
			}
		}
	}
	return len(data), nil
}

func splitParams(s string) []int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ";")
	result := make([]int, len(parts))
	for i, part := range parts {
		n, _ := strconv.Atoi(part)
		result[i] = n
	}
	return result
}

// decodeModifier extracts modifier flags from a CSI parameter.
// Terminals encode modifiers as 1 + bitmask (1=shift, 2=alt, 4=ctrl).
func decodeModifier(n int) Modifier {
	if n <= 1 {
		return 0
	}
	n--
	var mod Modifier
	if n&1 != 0 {
		mod |= ModShift
	}
	if n&2 != 0 {
		mod |= ModAlt
	}
	if n&4 != 0 {
		mod |= ModCtrl
	}
	return mod
}

func tildeKey(code int) Key {
	switch code {
	case 1:
		return KeyHome
	case 2:
		return KeyInsert
	case 3:
		return KeyDelete
	case 4:
		return KeyEnd
	case 5:
		return KeyPageUp
	case 6:
		return KeyPageDown
	case 11:
		return KeyF1
	case 12:
		return KeyF2
	case 13:
		return KeyF3
	case 14:
		return KeyF4
	case 15:
		return KeyF5
	case 17:
		return KeyF6
	case 18:
		return KeyF7
	case 19:
		return KeyF8
	case 20:
		return KeyF9
	case 21:
		return KeyF10
	case 23:
		return KeyF11
	case 24:
		return KeyF12
	default:
		return 0
	}
}
