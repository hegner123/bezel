package bezel

// Action represents the result of a key event processed by HandleEvent.
type Action uint8

const (
	ActionNone           Action = iota // No recognized binding.
	ActionQuit                         // Quit requested.
	ActionSubmit                       // Line submitted.
	ActionInsert                       // Character inserted.
	ActionPaste                        // Text pasted.
	ActionBackspace                    // Deleted before cursor.
	ActionDelete                       // Deleted at cursor.
	ActionLeft                         // Cursor left.
	ActionRight                        // Cursor right.
	ActionWordLeft                     // Cursor to previous word.
	ActionWordRight                    // Cursor to next word.
	ActionHome                         // Cursor to start.
	ActionEnd                          // Cursor to end.
	ActionDeleteToStart                // Cut to start of line.
	ActionDeleteToEnd                  // Cut to end of line.
	ActionDeleteWordBack               // Cut previous word.
	ActionHistoryPrev                  // Previous history entry.
	ActionHistoryNext                  // Next history entry.
	ActionNewline                      // Insert newline.
	ActionUp                           // Cursor up (line nav, then history).
	ActionDown                         // Cursor down (line nav, then history).
)

// KeyBind identifies a key combination for use as a KeyMap key.
// For KeyRune bindings (e.g., Ctrl+D), set Key=KeyRune and Ch='d'.
// For special keys (e.g., Enter), only Key and Mod are needed.
type KeyBind struct {
	Key Key
	Ch  rune
	Mod Modifier
}

// KeyMap maps key combinations to actions.
type KeyMap map[KeyBind]Action

// DefaultKeyMap returns a keymap with neovim insert-mode bindings.
// Movement uses arrow keys and Home/End. Ctrl+H, Ctrl+W, and Ctrl+U
// match vim's insert-mode deletion shortcuts.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		{Key: KeyEnter}:                       ActionSubmit,
		{Key: KeyRune, Ch: 'd', Mod: ModCtrl}: ActionQuit,

		{Key: KeyBackspace}:                   ActionBackspace,
		{Key: KeyDelete}:                      ActionDelete,
		{Key: KeyRune, Ch: 'h', Mod: ModCtrl}: ActionBackspace,
		{Key: KeyRune, Ch: 'w', Mod: ModCtrl}: ActionDeleteWordBack,
		{Key: KeyRune, Ch: 'u', Mod: ModCtrl}: ActionDeleteToStart,

		{Key: KeyLeft}:                ActionLeft,
		{Key: KeyRight}:               ActionRight,
		{Key: KeyLeft, Mod: ModCtrl}:  ActionWordLeft,
		{Key: KeyRight, Mod: ModCtrl}: ActionWordRight,
		{Key: KeyHome}:                ActionHome,
		{Key: KeyEnd}:                 ActionEnd,

		{Key: KeyEnter, Mod: ModAlt}: ActionNewline,

		{Key: KeyUp}:   ActionUp,
		{Key: KeyDown}: ActionDown,
	}
}

// EmacsKeyMap returns a keymap with emacs/readline bindings.
func EmacsKeyMap() KeyMap {
	return KeyMap{
		{Key: KeyEnter}:                       ActionSubmit,
		{Key: KeyRune, Ch: 'd', Mod: ModCtrl}: ActionQuit,

		{Key: KeyBackspace}:                   ActionBackspace,
		{Key: KeyDelete}:                      ActionDelete,
		{Key: KeyRune, Ch: 'h', Mod: ModCtrl}: ActionBackspace,
		{Key: KeyRune, Ch: 'w', Mod: ModCtrl}: ActionDeleteWordBack,
		{Key: KeyRune, Ch: 'u', Mod: ModCtrl}: ActionDeleteToStart,
		{Key: KeyRune, Ch: 'k', Mod: ModCtrl}: ActionDeleteToEnd,
		{Key: KeyBackspace, Mod: ModAlt}:      ActionDeleteWordBack,

		{Key: KeyLeft}:                        ActionLeft,
		{Key: KeyRight}:                       ActionRight,
		{Key: KeyLeft, Mod: ModCtrl}:          ActionWordLeft,
		{Key: KeyRight, Mod: ModCtrl}:         ActionWordRight,
		{Key: KeyHome}:                        ActionHome,
		{Key: KeyEnd}:                         ActionEnd,
		{Key: KeyRune, Ch: 'a', Mod: ModCtrl}: ActionHome,
		{Key: KeyRune, Ch: 'e', Mod: ModCtrl}: ActionEnd,
		{Key: KeyRune, Ch: 'b', Mod: ModCtrl}: ActionLeft,
		{Key: KeyRune, Ch: 'f', Mod: ModCtrl}: ActionRight,
		{Key: KeyRune, Ch: 'b', Mod: ModAlt}:  ActionWordLeft,
		{Key: KeyRune, Ch: 'f', Mod: ModAlt}:  ActionWordRight,

		{Key: KeyEnter, Mod: ModAlt}: ActionNewline,

		{Key: KeyUp}:                          ActionUp,
		{Key: KeyDown}:                        ActionDown,
		{Key: KeyRune, Ch: 'p', Mod: ModCtrl}: ActionHistoryPrev,
		{Key: KeyRune, Ch: 'n', Mod: ModCtrl}: ActionHistoryNext,
	}
}
