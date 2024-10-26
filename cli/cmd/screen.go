package cmd

import (
	"github.com/gdamore/tcell/v2"
)

// getTerminalHeight returns the current height of the terminal.
func getTerminalHeight() int {
	_, height := screen.Size()
	return height
}

// render prints the currently visible lines based on the scroll offset and highlights the selected line.
func render() {
	screen.Clear()
	_, height := screen.Size()
	maxLines := len(lines)
	linesLimit := height - 2
	for i := 0; i < linesLimit; i++ {
		lineIndex := i + scrollOffset
		if lineIndex >= maxLines {
			break
		}
		line := lines[lineIndex]
		style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault)
		if lineIndex == selectedLine {
			style = style.Background(tcell.ColorGrey)
		}
		if line.temperature > 90 {
			style = style.Foreground(tcell.ColorRed)
		}
		//lineInfo := strconv.Itoa(int(line.temperature))
		lineText := line.text // + " " + lineInfo

		// Print lineText
		for j, ch := range lineText {
			screen.SetContent(j, i, ch, nil, style)
		}

		// Print lineInfo in grey

		/*
			infoStyle := tcell.StyleDefault.Foreground(tcell.ColorGrey).Background(tcell.ColorDefault)
			for j, ch := range lineInfo {
				screen.SetContent(len(lineText)+1+j, i, ch, nil, infoStyle)
			}
		*/
	}
	// print a separator
	separator := "------------------------------------------------------------------"
	for i, ch := range separator {
		screen.SetContent(i, height-2, ch, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault))
	}

	// print status bar on screen
	// statusBar := "Press <esc> or <ctrl+c> to quit"
	line := lines[selectedLine]
	fileName := line.file.Name()
	lineTimeFmt := line.time.Format("2006-01-02 15:04:05")
	lineKind := line.kind
	statusBar := fileName + " | " + lineTimeFmt + " | " + lineKind.String()
	for i, ch := range statusBar {
		screen.SetContent(i, height-1, ch, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault))
	}

	screen.Show()
}

// handleScroll handles key events for scrolling up and down.
func handleScroll(quit chan struct{}) {
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				close(quit)
				return
			case tcell.KeyUp:
				if selectedLine > 0 {
					selectedLine--
					if selectedLine < scrollOffset {
						scrollOffset = selectedLine
					}
					render()
				}
			case tcell.KeyDown:
				if selectedLine < len(lines)-1 {
					selectedLine++
					if selectedLine >= scrollOffset+getTerminalHeight()-2 {
						scrollOffset++
					}
					render()
				}
			}
		case *tcell.EventResize:
			render()
		}
	}
}
