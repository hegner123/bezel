package main

import (
	"fmt"
	"os"

	"bezel"
)

func main() {
	c, err := bezel.New(os.Stdin, os.Stdout, 3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()

	var input []rune
	var history []string
	status := "ready"

	banner := func() {
		fmt.Println("Chrome Phase 3 — scroll region demo")
		fmt.Println("Type text, press Enter to submit.")
		fmt.Println("Submitted text appears here in the scroll region.")
		fmt.Println("")
		for _, line := range history {
			fmt.Println(line)
		}
	}

	redraw := func() {
		size := c.Size()
		c.Redraw(
			fmt.Sprintf("── %dx%d ── %s ", size.Cols, size.Rows, status),
			fmt.Sprintf("\033[1m>\033[0m %s\033[7m \033[0m", string(input)),
			"Enter submit | Backspace delete | Ctrl-D quit",
		)
	}

	banner()
	redraw()

	for ev := range c.Events() {
		switch ev.Type {
		case bezel.EventKey:
			switch {
			case ev.Key == bezel.KeyRune && ev.Ch == 'd' && ev.Mod == bezel.ModCtrl:
				return
			case ev.Key == bezel.KeyEnter:
				if len(input) > 0 {
					line := fmt.Sprintf("> %s", string(input))
					history = append(history, line)
					fmt.Println(line)
					status = fmt.Sprintf("submitted %d chars", len(input))
					input = input[:0]
				}
			case ev.Key == bezel.KeyBackspace:
				if len(input) > 0 {
					input = input[:len(input)-1]
				}
				status = "editing"
			case ev.Key == bezel.KeyRune && ev.Mod == 0:
				input = append(input, ev.Ch)
				status = "editing"
			case ev.Key == bezel.KeyRune && ev.Mod == bezel.ModAlt:
				status = fmt.Sprintf("Alt+%c", ev.Ch)
			case ev.Key != bezel.KeyRune:
				status = ev.Key.String()
				if ev.Mod != 0 {
					mod := ""
					if ev.Mod&bezel.ModCtrl != 0 {
						mod += "Ctrl+"
					}
					if ev.Mod&bezel.ModAlt != 0 {
						mod += "Alt+"
					}
					if ev.Mod&bezel.ModShift != 0 {
						mod += "Shift+"
					}
					status = mod + ev.Key.String()
				}
			}
		case bezel.EventPaste:
			input = append(input, []rune(ev.Text)...)
			status = fmt.Sprintf("pasted %d chars", len(ev.Text))
		case bezel.EventResize:
			status = "resized"
			banner()
		}
		redraw()
	}
}
