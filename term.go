package bezel

import (
	"fmt"
	"io"
	"os"
)

// RawState holds the original terminal state for restoration.
type RawState struct {
	fd  uintptr
	old termios
}

// EnableRaw puts the terminal into raw mode and returns the original state.
//
// Disables: echo, canonical mode, signal generation, extended input processing.
// Keeps: OPOST (output processing) so stdout from child processes renders normally.
// Sets: VMIN=1 (return after 1 byte), VTIME=0 (no timeout).
func EnableRaw(f *os.File) (*RawState, error) {
	fd := f.Fd()
	old, err := getTermios(fd)
	if err != nil {
		return nil, fmt.Errorf("enable raw mode: %w", err)
	}

	raw := old
	raw.Iflag &^= brkint | icrnl | inpck | istrip | ixon
	raw.Lflag &^= echo | icanon | iexten | isig
	raw.Cflag |= cs8
	raw.Cc[vmin] = 1
	raw.Cc[vtime] = 0

	if err := setTermios(fd, &raw); err != nil {
		return nil, fmt.Errorf("enable raw mode: %w", err)
	}
	return &RawState{fd: fd, old: old}, nil
}

// Restore returns the terminal to its original state.
func (s *RawState) Restore() error {
	return setTermios(s.fd, &s.old)
}

// Size represents terminal dimensions.
type Size struct {
	Rows uint16
	Cols uint16
}

// TermSize returns the current terminal dimensions.
func TermSize(f *os.File) (Size, error) {
	ws, err := getWinsize(f.Fd())
	if err != nil {
		return Size{}, fmt.Errorf("terminal size: %w", err)
	}
	return Size{Rows: ws.Rows, Cols: ws.Cols}, nil
}

// EnableBracketedPaste tells the terminal to wrap pasted text with
// escape sequences (\033[200~ ... \033[201~) so paste can be
// distinguished from typed input.
func EnableBracketedPaste(w io.Writer) error {
	_, err := w.Write([]byte("\033[?2004h"))
	return err
}

// DisableBracketedPaste turns off bracketed paste mode.
func DisableBracketedPaste(w io.Writer) error {
	_, err := w.Write([]byte("\033[?2004l"))
	return err
}
