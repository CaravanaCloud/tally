package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/nsf/termbox-go"
	"github.com/spf13/cobra"
)

// Global variables for managing the file content, scrolling, and selection.
var (
	fileContent  []string // Stores the lines of the file
	scrollOffset int      // Tracks the current scroll position
	selectedLine int      // Tracks the currently selected line
)

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// getTerminalHeight returns the current height of the terminal.
func getTerminalHeight() int {
	_, height := termbox.Size()
	return height
}

// printVisibleLines prints the currently visible lines based on the scroll offset and highlights the selected line.
func printVisibleLines() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	_, height := termbox.Size()
	maxLines := len(fileContent)

	// Calculate the visible range of lines based on scrollOffset and terminal height.
	for i := 0; i < height-1; i++ { // -1 to leave space for bottom border
		lineIndex := i + scrollOffset
		if lineIndex >= maxLines {
			break
		}
		if lineIndex == selectedLine {
			// Highlight the selected line with a different background color
			for j, ch := range fileContent[lineIndex] {
				termbox.SetCell(j, i, ch, termbox.ColorBlack, termbox.ColorGreen) // Black text on green background
			}
		} else {
			// Print normal lines
			for j, ch := range fileContent[lineIndex] {
				termbox.SetCell(j, i, ch, termbox.ColorWhite, termbox.ColorDefault) // Normal text
			}
		}
	}
	termbox.Flush()
}

// loadFile reads the file contents and stores it in the fileContent global variable.
func loadFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Split the content into lines and update fileContent.
	fileContent = splitLines(string(content))
	scrollOffset = len(fileContent) - getTerminalHeight() + 1 // Scroll to the end of the file
	if scrollOffset < 0 {
		scrollOffset = 0 // Ensure scrollOffset does not go negative
	}
	selectedLine = len(fileContent) - 1 // Set the selected line to the last line
	printVisibleLines()
	return nil
}

// splitLines splits the content into individual lines.
func splitLines(content string) []string {
	return strings.Split(content, "\n")
}

// watchFile watches the specified file for changes and reloads it when modified.
func watchFile(filePath string, quit chan struct{}) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Failed to create file watcher:", err)
	}
	defer watcher.Close()

	// Add the file to the watcher
	err = watcher.Add(filePath)
	if err != nil {
		log.Fatal("Failed to watch file:", err)
	}

	for {
		select {
		case <-quit:
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				if err := loadFile(filePath); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to reload file: %v\n", err)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("File watcher error:", err)
		}
	}
}

// handleScroll handles key events for scrolling up and down.
func handleScroll(quit chan struct{}) {
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc, termbox.KeyCtrlC: // Quit on Escape or Ctrl+C
				close(quit)
				return
			case termbox.KeyArrowUp: // Scroll up
				if selectedLine > 0 {
					selectedLine--
					if selectedLine < scrollOffset { // Adjust scroll if necessary
						scrollOffset = selectedLine
					}
					printVisibleLines()
				}
			case termbox.KeyArrowDown: // Scroll down
				if selectedLine < len(fileContent)-1 {
					selectedLine++
					if selectedLine >= scrollOffset+getTerminalHeight()-1 { // Adjust scroll if necessary
						scrollOffset++
					}
					printVisibleLines()
				}
			case termbox.KeyPgup: // Page up
				scrollOffset -= getTerminalHeight() - 1
				if scrollOffset < 0 {
					scrollOffset = 0
				}
				printVisibleLines()
			case termbox.KeyPgdn: // Page down
				scrollOffset += getTerminalHeight() - 1
				if scrollOffset+getTerminalHeight()-1 > len(fileContent) {
					scrollOffset = len(fileContent) - getTerminalHeight() + 1
					if scrollOffset < 0 {
						scrollOffset = 0
					}
				}
				printVisibleLines()
			}
		case termbox.EventResize: // Handle resize events
			printVisibleLines()
		}
	}
}

// rootCmd is the base command for the Cobra CLI.
var rootCmd = &cobra.Command{
	Use:   "tally [file]",
	Short: "Display file content and reload on changes",
	Long:  `Display file content and automatically reload when the file changes.`,
	Run:   runTally,
	Args:  cobra.MaximumNArgs(1), // Only accept up to one file argument
}

// runTally handles the main logic of the program.
func runTally(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Error: No file specified. Please provide a file path as an argument.")
		os.Exit(1) // Exit with an error code
	}

	filePath := args[0]
	if _, err := os.Stat(filePath); err != nil {
		log.Fatalf("Error: %s does not exist\n", filePath)
	}

	// Handle Ctrl+C to clean up and exit the program.
	quit := make(chan struct{})
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		close(quit)
	}()

	if err := termbox.Init(); err != nil {
		log.Fatal(err)
	}
	defer termbox.Close()

	if err := loadFile(filePath); err != nil {
		log.Fatalf("Failed to load file: %v", err)
	}

	// Start the file watcher and handle scrolling.
	go watchFile(filePath, quit)
	handleScroll(quit)

	fmt.Println("Program terminated.")
}

func init() {
	// Additional flags and configurations can be set here if needed.
}
