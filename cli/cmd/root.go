package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table table.Model
}

func generateFakeLogLine() LogLine {
	// Generating random timestamp
	timestamp := time.Now().Add(time.Duration(rand.Intn(10000)) * time.Minute)

	// Generating random log level
	logLevels := []string{"INFO", "WARN", "ERROR", "DEBUG"}
	logLevel := logLevels[rand.Intn(len(logLevels))]

	// Generating random message
	message := faker.Sentence()

	// Generating random source file
	sourceFile := fmt.Sprintf("file_%d.log", rand.Intn(10))

	return LogLine{
		SourceFile: sourceFile,
		Timestamp:  timestamp,
		Text:       fmt.Sprintf("%s: %s", logLevel, message),
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			selectedRow := m.table.SelectedRow()
			if selectedRow != nil {
				return m, tea.Batch(
					tea.Printf("Selected log line: %s", selectedRow[2]),
				)
			}
		case "up", "k":
			m.table.MoveUp(1)
		case "down", "j":
			m.table.MoveDown(1)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tally [file or directory]",
	Short: "Analyze a bundle of logfiles for issues",
	Long: `Analyze a bundle of logfiles for knows issues.`,
	Run: runTally,
	Args: cobra.MaximumNArgs(1),
}

func validatePath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("directories are not supported yet")
	}

	return nil
}

func displayLogLines(logLines []LogLine) {
	//TODO: Get current terminal width
	width := 120

	columns := []table.Column{
		{Title: "Source File", Width: 10},
		{Title: "Timestamp", Width: 15},
		{Title: "Log Text", Width: width - 25},
	}

	// Sort log lines by timestamp
	sort.Slice(logLines, func(i, j int) bool {
		return logLines[i].Timestamp.Before(logLines[j].Timestamp)
	})

	//TODO: Print only file name, time 
	var rows []table.Row
	for _, logLine := range logLines {
		rows = append(rows, table.Row{
			logLine.SourceFile,
			logLine.Timestamp.Format(time.RFC3339),
			logLine.Text,
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{t}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}



func runTally(cmd *cobra.Command, args []string) {
	var path string
	if len(args) == 0 {
		// Default to the current directory if no argument is provided
		path = "."
	} else {
		path = args[0]
	}

	if err := validatePath(path); err != nil {
		fmt.Println("Error:", err)
		return
	}

	lines, err := loadLinesFromFile(path)
	if err != nil {
		fmt.Println("Error loading file:", err)
		return
	}

	// Process and display lines
	displayLogLines(lines)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tally.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
