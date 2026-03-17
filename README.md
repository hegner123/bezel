# bezel

Go library for terminal REPL chrome. Pins a fixed UI (prompt, status bar, hints) at the bottom of the terminal while stdout flows freely above it. Zero dependencies beyond the standard library.

Built for agentic CLI tools where you need a persistent input area but don't want to take over the screen. Uses ANSI scroll regions — the same technique as Claude Code.

## Install

```
go get github.com/hegner123/bezel@latest
```

Requires Go 1.22+. Supports macOS and Linux (amd64, arm64).

## Quick start

```go
b, err := bezel.New(os.Stdin, os.Stdout, 3) // 3 rows of chrome at the bottom
if err != nil {
    log.Fatal(err)
}
defer b.Close()

var ed bezel.LineEditor
var hist bezel.History
km := bezel.DefaultKeyMap()

for ev := range b.Events() {
    action, text := ed.HandleEvent(ev, km, &hist)
    switch action {
    case bezel.ActionQuit:
        return
    case bezel.ActionSubmit:
        hist.Add(text)
        b.CursorToScroll()
        fmt.Println(text) // goes to scroll region
    }
    b.RedrawPrompt(1, 2+ed.Pos(), "status", "> "+ed.String(), "hints")
}
```

## How it works

```
┌──────────────────────────────────────┐
│ Previous terminal output preserved   │
│ Child process stdout lands here      │  scroll region
│ LLM streaming tokens land here       │  (rows 1 to N-3)
│ fmt.Println() lands here             │
├──────────────────────────────────────┤
│ ── 80x24 ── thinking...             │  bezel row 0 (status)
│ > user input here█                  │  bezel row 1 (prompt)
│ Enter submit | Ctrl-D quit          │  bezel row 2 (hints)
└──────────────────────────────────────┘
```

The scroll region is standard terminal scrolling — programs write to stdout normally and it just works. The bezel is redrawn via ANSI escape sequences as a single atomic write to prevent tearing.

Previous terminal history (before your tool launched) is preserved and scrollable.

## Bezel

`Bezel` manages the scroll region, raw terminal mode, bracketed paste, SIGWINCH, and the merged event channel.

```go
// Create. Height is the number of chrome rows at the bottom.
b, err := bezel.New(os.Stdin, os.Stdout, 3)
defer b.Close()

// Read events (keyboard, paste, resize).
for ev := range b.Events() { ... }

// Current terminal dimensions.
size := b.Size() // Size{Rows, Cols}
```

### Redraw methods

Two methods for drawing the bezel, depending on where the cursor should end up:

```go
// Redraw: cursor stays in the scroll region.
// Use during output streaming — stdout writes continue uninterrupted.
b.Redraw("status line", "> partial output...", "hints")

// RedrawPrompt: cursor moves to (row, col) within the bezel.
// Use during interactive input — user sees a cursor at the prompt.
// row is 0-indexed from top of bezel, col is 0-indexed from left.
b.RedrawPrompt(1, cursorCol, "status", "> input", "hints")
```

### Switching between modes

An agentic tool alternates between streaming output and waiting for input:

```go
// Tool output phase: cursor in scroll region.
b.Redraw("thinking...", "> ", "Ctrl-C cancel")
streamLLMOutput(os.Stdout) // tokens flow to scroll region

// Input phase: cursor in bezel.
b.RedrawPrompt(1, 2+ed.Pos(), "ready", "> "+ed.String(), "Enter submit")

// Before writing to stdout when cursor is in the bezel:
b.CursorToScroll()
fmt.Println("output") // lands in scroll region
b.RedrawPrompt(...)    // cursor back to bezel
```

### Resize

On terminal resize, the screen is cleared and an `EventResize` is delivered. Re-emit any scroll region content in your handler:

```go
if ev.Type == bezel.EventResize {
    // Re-print whatever should be visible.
    for _, line := range outputHistory {
        fmt.Println(line)
    }
    redraw()
}
```

## Events

A single channel delivers all input. Events are parsed from raw terminal bytes — escape sequences become structured types.

```go
type Event struct {
    Type EventType // EventKey, EventPaste, EventResize, EventUnknown
    Key  Key       // Which key (KeyRune, KeyEnter, KeyUp, KeyF1, ...)
    Ch   rune      // The character for KeyRune events
    Mod  Modifier  // ModCtrl, ModAlt, ModShift (bitfield)
    Text string    // Paste content for EventPaste
    Raw  []byte    // Original bytes, always set
}
```

### Event types

| Type | When |
|------|------|
| `EventKey` | Any key press. Check `Key`, `Ch`, `Mod`. |
| `EventPaste` | Bracketed paste. Full text in `Text`. |
| `EventResize` | Terminal resized. Call `Size()` for new dimensions. |
| `EventUnknown` | Unrecognized escape sequence. Raw bytes in `Raw`. |

### Keys

Special keys: `KeyEnter`, `KeyTab`, `KeyBackspace`, `KeyEscape`, `KeyUp`, `KeyDown`, `KeyLeft`, `KeyRight`, `KeyHome`, `KeyEnd`, `KeyDelete`, `KeyInsert`, `KeyPageUp`, `KeyPageDown`, `KeyF1`–`KeyF12`.

For printable characters, `Key == KeyRune` and `Ch` holds the rune. Ctrl+letter shows as `Key=KeyRune, Ch='c', Mod=ModCtrl`.

Modified special keys work: `Key=KeyRight, Mod=ModCtrl` for Ctrl+Right, etc.

## Line editor

`LineEditor` manages an editable line of text with cursor position. Zero value is ready to use.

```go
var ed bezel.LineEditor
```

### Direct methods

For manual control, call methods directly:

```go
ed.Insert('x')        // insert at cursor
ed.InsertString("hi") // insert string (paste)
ed.Backspace()         // delete before cursor
ed.Delete()            // delete at cursor
ed.Left()              // cursor left
ed.Right()             // cursor right
ed.Home()              // cursor to start
ed.End()               // cursor to end
ed.WordLeft()          // cursor to previous word boundary
ed.WordRight()         // cursor past next word
ed.DeleteToStart()     // cut to start of line (Ctrl-U)
ed.DeleteToEnd()       // cut to end of line (Ctrl-K)
ed.DeleteWordBack()    // cut previous word (Ctrl-W)
text := ed.Submit()    // return content, reset editor
ed.Set("preset")       // replace content, cursor to end
ed.Clear()             // empty the editor

ed.String()            // current content
ed.Pos()               // cursor position (runes from start)
ed.Len()               // content length in runes
ed.Empty()             // true if no content
```

### HandleEvent

For the common case, `HandleEvent` maps events to editor actions via a configurable keymap:

```go
action, text := ed.HandleEvent(ev, km, &hist)
```

Returns the `Action` taken and, for `ActionSubmit`/`ActionPaste`, the relevant text. Pass `nil` for `hist` if history is not needed.

## Keymaps

`KeyMap` maps key combinations to actions. `DefaultKeyMap()` provides standard terminal and emacs bindings.

### Default bindings

| Key | Action |
|-----|--------|
| Enter | Submit |
| Ctrl-D | Quit |
| Backspace, Ctrl-H | Backspace |
| Delete | Delete |
| Left, Ctrl-B | Cursor left |
| Right, Ctrl-F | Cursor right |
| Ctrl-Left, Alt-B | Word left |
| Ctrl-Right, Alt-F | Word right |
| Home, Ctrl-A | Home |
| End, Ctrl-E | End |
| Ctrl-U | Cut to start |
| Ctrl-K | Cut to end |
| Ctrl-W, Alt-Backspace | Cut word back |
| Up, Ctrl-P | History previous |
| Down, Ctrl-N | History next |

### Customizing

```go
km := bezel.DefaultKeyMap()

// Change quit to Ctrl-C.
delete(km, bezel.KeyBind{Key: bezel.KeyRune, Ch: 'd', Mod: bezel.ModCtrl})
km[bezel.KeyBind{Key: bezel.KeyRune, Ch: 'c', Mod: bezel.ModCtrl}] = bezel.ActionQuit

// Add Ctrl-L (handle before HandleEvent for custom behavior).
// Or bind it to an existing action:
km[bezel.KeyBind{Key: bezel.KeyRune, Ch: 'l', Mod: bezel.ModCtrl}] = bezel.ActionDeleteToStart

// Remove a binding.
delete(km, bezel.KeyBind{Key: bezel.KeyRune, Ch: 'k', Mod: bezel.ModCtrl})
```

For actions beyond the built-in set, handle the event before calling `HandleEvent`:

```go
for ev := range b.Events() {
    // Custom bindings first.
    if ev.Type == bezel.EventKey && ev.Key == bezel.KeyRune && ev.Ch == 'l' && ev.Mod == bezel.ModCtrl {
        clearScreen()
        continue
    }

    // Then standard editor handling.
    action, text := ed.HandleEvent(ev, km, &hist)
    ...
}
```

### Actions

| Action | Meaning |
|--------|---------|
| `ActionNone` | Unrecognized key, no change |
| `ActionQuit` | Quit requested |
| `ActionSubmit` | Line submitted (text in return value) |
| `ActionInsert` | Character inserted |
| `ActionPaste` | Text pasted (text in return value) |
| `ActionBackspace` | Deleted before cursor |
| `ActionDelete` | Deleted at cursor |
| `ActionLeft`, `ActionRight` | Cursor moved |
| `ActionWordLeft`, `ActionWordRight` | Cursor jumped by word |
| `ActionHome`, `ActionEnd` | Cursor jumped to boundary |
| `ActionDeleteToStart`, `ActionDeleteToEnd` | Line cut |
| `ActionDeleteWordBack` | Word cut |
| `ActionHistoryPrev`, `ActionHistoryNext` | History navigation |

## History

`History` stores submitted lines and supports Up/Down navigation with draft preservation. Zero value is ready to use.

```go
var hist bezel.History

// Add entries (caller decides what enters history).
hist.Add("command")

// Navigation is handled automatically by HandleEvent.
// Or manually:
text, ok := hist.Prev(currentInput) // saves current input as draft
text, ok = hist.Next()              // returns draft when past newest
hist.Reset()                        // stop navigating

// For persistence or display:
hist.Entries() // []string, oldest to newest
hist.Len()     // number of entries
```

Consecutive duplicates and empty strings are automatically skipped on `Add`.

When the user presses Up, their current input is saved as a draft. Navigating back down past the newest entry restores it.

## Agentic tool pattern

Complete pattern for a tool that runs LLM calls and streams output:

```go
func main() {
    b, _ := bezel.New(os.Stdin, os.Stdout, 3)
    defer b.Close()

    var ed bezel.LineEditor
    var hist bezel.History
    km := bezel.DefaultKeyMap()

    redraw := func(status string) {
        b.RedrawPrompt(1, 2+ed.Pos(),
            status,
            "> "+ed.String(),
            "Enter send | Ctrl-D quit",
        )
    }

    redraw("ready")

    for ev := range b.Events() {
        if ev.Type == bezel.EventResize {
            redraw("ready")
            continue
        }

        action, text := ed.HandleEvent(ev, km, &hist)
        switch action {
        case bezel.ActionQuit:
            return
        case bezel.ActionSubmit:
            hist.Add(text)

            // Print the user's message to scroll region.
            b.CursorToScroll()
            fmt.Printf("You: %s\n", text)

            // Stream LLM response. Cursor is in scroll region
            // so streamed tokens render correctly.
            b.Redraw("thinking...", "> ", "Ctrl-C cancel")
            response := callLLM(text)
            fmt.Printf("AI: %s\n", response)

            redraw("ready")
            continue
        case bezel.ActionNone:
            continue
        }
        redraw("editing")
    }
}
```

## Lower-level API

The high-level `Bezel` type composes these primitives, which are also exported:

```go
// Raw terminal mode.
state, err := bezel.EnableRaw(os.Stdin)
defer state.Restore()

// Terminal dimensions.
size, err := bezel.TermSize(os.Stdin)

// Bracketed paste mode.
bezel.EnableBracketedPaste(os.Stdout)
defer bezel.DisableBracketedPaste(os.Stdout)

// Parsed input event channel.
ctx, cancel := context.WithCancel(context.Background())
events := bezel.ReadInput(ctx, os.Stdin)
for ev := range events { ... }
```

Use these if you need raw terminal control without the scroll region chrome.

## Platform

macOS and Linux. Uses `ioctl` syscalls via `syscall.Syscall`. No cgo.

- **macOS:** `TIOCGETA`/`TIOCSETA`/`TIOCGWINSZ`
- **Linux:** `TCGETS`/`TCSETS`/`TIOCGWINSZ`

## License

MIT — see [LICENSE](LICENSE).
