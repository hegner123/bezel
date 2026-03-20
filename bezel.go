package bezel

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Bezel manages a terminal scroll region with fixed bezel rows at the bottom.
// stdout flows naturally in the scroll region (top area). The bezel area
// (bottom N rows) is redrawn via Redraw. The real terminal cursor is hidden
// permanently; use LineEditor.StringWithCursor to render a pseudo cursor (█)
// as part of the bezel content. Events from keyboard input and terminal
// resize are delivered on a single channel.
type Bezel struct {
	in     *os.File
	out    *os.File
	raw    *RawState
	height int

	mu   sync.Mutex
	size Size

	writeMu sync.Mutex

	merged chan Event
	sigCh  chan os.Signal
	cancel context.CancelFunc
}

// New creates a Bezel with bezelHeight initial rows at the bottom of the
// terminal. The height adjusts dynamically on each Redraw call to fit the
// number of lines provided. It enters raw mode, enables bracketed paste,
// sets up the scroll region, and starts reading input. Call Close to
// restore the terminal.
func New(in, out *os.File, bezelHeight int) (*Bezel, error) {
	size, err := TermSize(in)
	if err != nil {
		return nil, fmt.Errorf("bezel: %w", err)
	}

	raw, err := EnableRaw(in)
	if err != nil {
		return nil, fmt.Errorf("bezel: %w", err)
	}

	if err := EnableBracketedPaste(out); err != nil {
		raw.Restore()
		return nil, fmt.Errorf("bezel: enable paste: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := &Bezel{
		in:     in,
		out:    out,
		raw:    raw,
		height: bezelHeight,
		size:   size,
		merged: make(chan Event, 64),
		sigCh:  make(chan os.Signal, 1),
		cancel: cancel,
	}

	c.initScrollRegion()

	signal.Notify(c.sigCh, syscall.SIGWINCH)
	inputCh := ReadInput(ctx, in)
	go c.run(ctx, inputCh)

	return c, nil
}

// initScrollRegion sets the scroll region and clears only the bezel rows.
// Existing terminal content in the scroll region is preserved.
func (c *Bezel) initScrollRegion() {
	c.mu.Lock()
	size := c.size
	c.mu.Unlock()

	sb := scrollBottom(int(size.Rows), c.height)

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	var buf bytes.Buffer
	buf.WriteString("\0337")            // DECSC: save cursor position
	fmt.Fprintf(&buf, "\033[1;%dr", sb) // set scroll region

	// Clear only the bezel rows.
	for i := range c.height {
		row := sb + i + 1
		fmt.Fprintf(&buf, "\033[%d;1H\033[2K", row)
	}

	buf.WriteString("\0338")     // DECRC: restore cursor to original position
	buf.WriteString("\033[?25l") // hide cursor permanently
	c.out.Write(buf.Bytes())
}

// Events returns the channel of input and resize events.
// The channel is closed when Close is called or input ends.
func (c *Bezel) Events() <-chan Event {
	return c.merged
}

// Size returns the current terminal dimensions.
func (c *Bezel) Size() Size {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.size
}

// Redraw clears and redraws the bezel with the given lines.
// The bezel height adjusts dynamically to fit len(lines), resizing
// the scroll region as needed. When the height changes, the cursor
// is parked at the bottom of the new scroll region; otherwise it is
// restored to its previous position.
// Use LineEditor.Visual to get wrapped rows with the pseudo cursor
// embedded.
func (c *Bezel) Redraw(lines ...string) {
	c.mu.Lock()
	size := c.size
	c.mu.Unlock()

	newHeight := len(lines)
	maxHeight := int(size.Rows) - 1
	if newHeight > maxHeight {
		newHeight = maxHeight
	}
	if newHeight < 1 {
		newHeight = 1
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	oldHeight := c.height
	oldSB := scrollBottom(int(size.Rows), oldHeight)
	newSB := scrollBottom(int(size.Rows), newHeight)
	heightChanged := newHeight != oldHeight

	var buf bytes.Buffer
	buf.WriteString("\0337") // DECSC: save cursor position

	if heightChanged {
		// Clear old bezel rows before resizing.
		for i := range oldHeight {
			row := oldSB + i + 1
			fmt.Fprintf(&buf, "\033[%d;1H\033[2K", row)
		}
		// Set new scroll region.
		fmt.Fprintf(&buf, "\033[1;%dr", newSB)
		c.height = newHeight
	}

	// Draw bezel rows.
	for i := range newHeight {
		row := newSB + i + 1
		fmt.Fprintf(&buf, "\033[%d;1H\033[2K", row)
		if i < len(lines) {
			buf.WriteString(lines[i])
		}
	}

	if heightChanged {
		// Cursor may be outside new scroll region. Park at bottom.
		fmt.Fprintf(&buf, "\033[%d;1H", newSB)
	} else {
		buf.WriteString("\0338") // DECRC: restore cursor to scroll region
	}
	c.out.Write(buf.Bytes())
}

// Close restores the terminal to its original state.
// After Close returns, the Events channel is closed.
func (c *Bezel) Close() error {
	c.cancel()
	signal.Stop(c.sigCh)

	c.mu.Lock()
	size := c.size
	c.mu.Unlock()

	c.writeMu.Lock()
	var buf bytes.Buffer
	buf.WriteString("\033[r")                  // reset scroll region
	fmt.Fprintf(&buf, "\033[%d;1H", size.Rows) // move to last row
	buf.WriteString("\033[2K\n")               // clear it, newline
	buf.WriteString("\033[?25h")               // ensure cursor visible
	c.out.Write(buf.Bytes())
	c.writeMu.Unlock()

	DisableBracketedPaste(c.out)
	return c.raw.Restore()
}

func (c *Bezel) run(ctx context.Context, inputCh <-chan Event) {
	defer close(c.merged)

	for {
		select {
		case ev, ok := <-inputCh:
			if !ok {
				return
			}
			select {
			case c.merged <- ev:
			case <-ctx.Done():
				return
			}
		case <-c.sigCh:
			newSize, err := TermSize(c.in)
			if err != nil {
				continue
			}
			c.mu.Lock()
			c.size = newSize
			c.mu.Unlock()

			c.resetScrollRegion()

			select {
			case c.merged <- Event{Type: EventResize}:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// resetScrollRegion clears the entire display and re-establishes the scroll
// region. Used on terminal resize where content reflow makes preservation
// unreliable. The user's EventResize handler should re-emit any content.
func (c *Bezel) resetScrollRegion() {
	c.mu.Lock()
	size := c.size
	c.mu.Unlock()

	sb := scrollBottom(int(size.Rows), c.height)

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	var buf bytes.Buffer
	buf.WriteString("\033[r")           // reset scroll region to full screen
	buf.WriteString("\033[2J")          // clear entire display
	fmt.Fprintf(&buf, "\033[1;%dr", sb) // set new scroll region
	buf.WriteString("\033[1;1H")        // cursor at top of scroll region
	c.out.Write(buf.Bytes())
}

func scrollBottom(totalRows, bezelHeight int) int {
	sb := totalRows - bezelHeight
	if sb < 1 {
		return 1
	}
	return sb
}
