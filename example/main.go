package main

import (
	"fmt"
	"os"

	"github.com/hegner123/bezel"
)

const prompt = "> "

func main() {
	b, err := bezel.New(os.Stdin, os.Stdout, 3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer b.Close()

	var ed bezel.LineEditor
	var hist bezel.History
	status := "ready"
	km := bezel.DefaultKeyMap()

	banner := func() {
		fmt.Println("Bezel — line editor with history")
		fmt.Println("Up/Down navigate history")
		fmt.Println("")
		for _, entry := range hist.Entries() {
			fmt.Println(prompt + entry)
		}
	}

	redraw := func() {
		size := b.Size()
		v := ed.Visual(int(size.Cols), []string{prompt})
		lines := []string{fmt.Sprintf("── %dx%d ── %s ", size.Cols, size.Rows, status)}
		lines = append(lines, v.Rows...)
		lines = append(lines, "Enter submit | Up/Down history | Ctrl-C quit")
		b.Redraw(lines...)
	}

	banner()
	redraw()

	for ev := range b.Events() {
		if ev.Type == bezel.EventResize {
			status = "resized"
			banner()
			redraw()
			continue
		}

		action, text := ed.HandleEvent(ev, km, &hist)
		switch action {
		case bezel.ActionQuit:
			return
		case bezel.ActionSubmit:
			hist.Add(text)
			fmt.Println(prompt + text)
			status = fmt.Sprintf("submitted %d chars", len([]rune(text)))
		case bezel.ActionPaste:
			status = fmt.Sprintf("pasted %d chars", len([]rune(text)))
		case bezel.ActionHistoryPrev, bezel.ActionHistoryNext:
			status = "history"
		case bezel.ActionDeleteToStart:
			status = "cut to start"
		case bezel.ActionDeleteToEnd:
			status = "cut to end"
		case bezel.ActionDeleteWordBack:
			status = "cut word"
		case bezel.ActionNone:
			continue
		default:
			status = "editing"
		}
		redraw()
	}
}
