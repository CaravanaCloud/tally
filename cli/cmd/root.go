package cmd

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/spf13/cobra"
)

type quitChannel chan struct{}

type LogLine struct {
	text string
	time time.Time
}

const (
	logFileExtension = ".log"
	defaultDir       = "."
)

// Global variables
var (
	lines        []LogLine
	scrollOffset int
	selectedLine int
	mu           sync.Mutex // Mutex for concurrent access
	screen       tcell.Screen
)

// rootCmd is the base command for the Cobra CLI.
var rootCmd = &cobra.Command{
	Use:   "tally [directory]",
	Short: "Display log files in a directory and reload on changes",
	Run:   run,
	Args:  cobra.MaximumNArgs(1),
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
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

	handleScroll(quitChannel)
	log.Printf("Tally terminated")
}

func init() {}
