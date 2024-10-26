package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell/v2"
	"github.com/spf13/cobra"
)

const (
	logFileExtension = ".log"
	defaultDir       = "."
)

// Global variables
var (
	fileContent  []string
	scrollOffset int
	selectedLine int
	mu           sync.Mutex // Mutex for concurrent access
	screen       tcell.Screen
)

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// getTerminalHeight returns the current height of the terminal.
func getTerminalHeight() int {
	_, height := screen.Size()
	return height
}

// printVisibleLines prints the currently visible lines based on the scroll offset and highlights the selected line.
func printVisibleLines() {
	screen.Clear()
	_, height := screen.Size()
	maxLines := len(fileContent)

	for i := 0; i < height-1; i++ {
		lineIndex := i + scrollOffset
		if lineIndex >= maxLines {
			break
		}
		line := fileContent[lineIndex]
		style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault)
		if lineIndex == selectedLine {
			style = style.Background(tcell.ColorGreen)
		}
		for j, ch := range line {
			screen.SetContent(j, i, ch, nil, style)
		}
	}
	screen.Show()
}

// loadFilesInDirectory reads all log files from the specified directory recursively.
func loadFilesInDirectory(dirPath string) error {
	mu.Lock() // Lock to prevent concurrent access
	defer mu.Unlock()

	fileContent = nil

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), logFileExtension) {
			content, err := os.ReadFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read file %s: %v\n", path, err)
				return nil // Continue walking even if a file fails to read
			}
			lines := splitLines(string(content))
			fileContent = append(fileContent, lines...)
			fmt.Printf("Loaded file: %s (%d lines)\n", path, len(lines)) // Debug output
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error gathering log files: %w", err)
	}

	if len(fileContent) == 0 {
		fmt.Println("No log files found in directory.")
	}
	scrollOffset = len(fileContent) - getTerminalHeight() + 1
	if scrollOffset < 0 {
		scrollOffset = 0
	}
	selectedLine = len(fileContent) - 1
	printVisibleLines()
	return nil
}

// splitLines splits the content into individual lines.
func splitLines(content string) []string {
	return strings.Split(strings.TrimSpace(content), "\n")
}

// watchFiles watches the specified directory for changes and reloads log files when modified.
func watchFiles(dirPath string, quit chan struct{}) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Failed to create file watcher:", err)
	}
	defer watcher.Close()

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), logFileExtension) {
			if err := watcher.Add(path); err != nil {
				log.Printf("Failed to watch file %s: %v", path, err)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal("Failed to watch files:", err)
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
				fmt.Printf("File modified: %s\n", event.Name) // Feedback on file change
				loadFilesInDirectory(dirPath)
			}
		case err := <-watcher.Errors:
			log.Println("File watcher error:", err)
		}
	}
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
					printVisibleLines()
				}
			case tcell.KeyDown:
				if selectedLine < len(fileContent)-1 {
					selectedLine++
					if selectedLine >= scrollOffset+getTerminalHeight()-1 {
						scrollOffset++
					}
					printVisibleLines()
				}
			}
		case *tcell.EventResize:
			printVisibleLines()
		}
	}
}

// rootCmd is the base command for the Cobra CLI.
var rootCmd = &cobra.Command{
	Use:   "tally [directory]",
	Short: "Display log files in a directory and reload on changes",
	Run:   run,
	Args:  cobra.MaximumNArgs(1),
}

type settings struct {
	path string
}

func initSettings(args []string) settings {
	path := defaultDir
	if len(args) > 0 {
		path = args[0]
	}
	if _, err := os.Stat(path); err != nil {
		log.Fatalf("Error: %s does not exist\n", path)
	}
	return settings{path: path}
}

type quitChannel chan struct{}

func initSignals() quitChannel {
	quitChannel := make(quitChannel)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		close(quitChannel)
	}()
	return quitChannel
}

func run(cmd *cobra.Command, args []string) {
	quitChannel := initSignals()
	settings := initSettings(args)

	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err = screen.Init(); err != nil {
		log.Fatal(err)
	}
	defer screen.Fini()

	if err := loadFilesInDirectory(settings.path); err != nil {
		log.Fatalf("Failed to load files: %v", err)
	}

	go watchFiles(settings.path, quitChannel)
	handleScroll(quitChannel)
	log.Printf("Tally terminated")
}

func init() {}
