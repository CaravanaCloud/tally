package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nsf/termbox-go"
	"github.com/spf13/cobra"
)

// getTerminalHeight returns the current height of the terminal.
func getTerminalHeight() int {
	_, height := termbox.Size()
	return height
}

// printLines prints line numbers up to the current terminal height.
func printLines() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	height := getTerminalHeight()
	for i := 1; i <= height; i++ {
		fmt.Fprintf(os.Stdout, "Line %d\n", i)
	}
	termbox.Flush()	
}

// waitForResize continuously listens for resize events.
func waitForResize(quit chan struct{}) {
	for {
		select {
		case <-quit:
			return
		default:
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventResize:
				printLines()
			case termbox.EventInterrupt:
				return
			}
		}
	}
}

func main() {
	if err := termbox.Init(); err != nil {
		log.Fatal(err)
	}
	defer termbox.Close()

	// Handle Ctrl+C to clean up and exit the program.
	quit := make(chan struct{})
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		close(quit)
	}()

	printLines()
	waitForResize(quit)
}

var rootCmd = &cobra.Command{
	Use:   "tally [file or directory]",
	Short: "Analyze a bundle of logfiles for issues",
	Long:  `Analyze a bundle of logfiles for known issues.`,
	Run:   runTally,
	Args:  cobra.MaximumNArgs(1),
}

func runTally(cmd *cobra.Command, args []string) {
	main()
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
