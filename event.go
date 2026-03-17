package bezel

// EventType classifies an input event.
type EventType uint8

const (
	// EventUnknown indicates an unrecognized byte sequence.
	EventUnknown EventType = iota
	// EventKey indicates a key press (character, special key, or control).
	EventKey
	// EventPaste indicates bracketed paste content.
	EventPaste
	// EventResize indicates the terminal was resized. Call Size() for new dimensions.
	EventResize
)

// Key identifies a key. For printable characters, Key is KeyRune
// and the character is in Event.Ch.
type Key uint16

const (
	KeyRune Key = iota // Printable character; see Event.Ch.
	KeyEnter
	KeyTab
	KeyBackspace
	KeyEscape
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyDelete
	KeyInsert
	KeyPageUp
	KeyPageDown
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
)

// Modifier flags for key events.
type Modifier uint8

const (
	ModShift Modifier = 1 << iota
	ModAlt
	ModCtrl
)

// Event represents a parsed terminal input event.
type Event struct {
	Type EventType
	Key  Key      // Which key (for EventKey).
	Ch   rune     // Character (for KeyRune, or the letter for Ctrl+letter).
	Mod  Modifier // Active modifiers.
	Text string   // Paste content (for EventPaste).
	Raw  []byte   // Original bytes, always set.
}

func (k Key) String() string {
	switch k {
	case KeyRune:
		return "Rune"
	case KeyEnter:
		return "Enter"
	case KeyTab:
		return "Tab"
	case KeyBackspace:
		return "Backspace"
	case KeyEscape:
		return "Escape"
	case KeyUp:
		return "Up"
	case KeyDown:
		return "Down"
	case KeyLeft:
		return "Left"
	case KeyRight:
		return "Right"
	case KeyHome:
		return "Home"
	case KeyEnd:
		return "End"
	case KeyDelete:
		return "Delete"
	case KeyInsert:
		return "Insert"
	case KeyPageUp:
		return "PageUp"
	case KeyPageDown:
		return "PageDown"
	case KeyF1:
		return "F1"
	case KeyF2:
		return "F2"
	case KeyF3:
		return "F3"
	case KeyF4:
		return "F4"
	case KeyF5:
		return "F5"
	case KeyF6:
		return "F6"
	case KeyF7:
		return "F7"
	case KeyF8:
		return "F8"
	case KeyF9:
		return "F9"
	case KeyF10:
		return "F10"
	case KeyF11:
		return "F11"
	case KeyF12:
		return "F12"
	default:
		return "Unknown"
	}
}
