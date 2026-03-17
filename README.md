# bezel

A Go library for building terminal REPL interfaces with fixed chrome around free-flowing stdout. Zero dependencies beyond the standard library.

Bezel splits your terminal into two areas: a **scroll region** where stdout flows naturally (child process output, streaming text, logs), and a fixed **bezel** at the bottom for your UI chrome (prompt, status bar, key hints). This is the same technique used by Claude Code and similar agentic CLI tools.

## Features

- Scroll region management with fixed chrome rows at the bottom of the terminal
- Raw terminal mode via direct `ioctl` syscalls (no cgo, no external packages)
- Escape sequence parser: arrow keys, function keys, Home/End/Delete/Insert, modifiers (Ctrl, Alt, Shift)
- UTF-8 multi-byte character support
- Bracketed paste detection
- Terminal resize handling (SIGWINCH) with clean redraw
- Atomic chrome writes to prevent interleaving with concurrent stdout
- Single merged event channel for keyboard input, paste, and resize events

## Installation

```
go get bezel
```

## Usage

```go
package main

import (
    "fmt"
    "os"

    "bezel"
)

func main() {
    b, err := bezel.New(os.Stdin, os.Stdout, 3) // 3 chrome rows at bottom
    if err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
    defer b.Close()

    b.Redraw(
        "── status bar ──",
        "> ",
        "Ctrl-D quit",
    )

    for ev := range b.Events() {
        switch ev.Type {
        case bezel.EventKey:
            if ev.Key == bezel.KeyRune && ev.Ch == 'd' && ev.Mod == bezel.ModCtrl {
                return
            }
            // stdout goes to the scroll region automatically
            fmt.Printf("key: %s\n", ev.Key)
        case bezel.EventPaste:
            fmt.Printf("pasted: %s\n", ev.Text)
        case bezel.EventResize:
            // re-emit scroll region content if needed
        }
        b.Redraw("── status ──", "> ", "Ctrl-D quit")
    }
}
```

## API

```
bezel.New(in, out *os.File, height int) (*Bezel, error)
(*Bezel).Events() <-chan Event
(*Bezel).Redraw(lines ...string)
(*Bezel).Size() Size
(*Bezel).Close() error
```

Lower-level functions are also exported for custom use:

```
bezel.EnableRaw(f *os.File) (*RawState, error)
bezel.ReadInput(ctx context.Context, input io.Reader) <-chan Event
bezel.TermSize(f *os.File) (Size, error)
bezel.EnableBracketedPaste(w io.Writer) error
```

## License

MIT — see [LICENSE](LICENSE).
