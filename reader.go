package bezel

import (
	"context"
	"io"
)

// ReadInput reads from input, parses escape sequences and UTF-8, and returns
// a channel of structured Events. The channel is closed when the context is
// canceled or input returns an error (including EOF).
//
// The goroutine blocks on Read. To unblock it on shutdown, close the
// underlying file descriptor from another goroutine.
func ReadInput(ctx context.Context, input io.Reader) <-chan Event {
	ch := make(chan Event, 64)
	go func() {
		defer close(ch)
		p := newParser()
		buf := make([]byte, 256)
		for {
			n, err := input.Read(buf)
			if n > 0 {
				events := p.Parse(buf[:n])
				for _, ev := range events {
					select {
					case ch <- ev:
					case <-ctx.Done():
						return
					}
				}
			}
			if err != nil {
				return
			}
		}
	}()
	return ch
}
