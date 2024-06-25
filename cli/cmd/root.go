package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table  table.Model
	detail string
}

func (m model) Init() tea.Cmd {
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
		case "up", "k":
			m.table.MoveUp(1)
		case "down", "j":
			m.table.MoveDown(1)
		}
		// Update the detail information when the selection changes
		selectedRow := m.table.SelectedRow()
		if selectedRow != nil {
			m.detail = fmt.Sprintf("Source File: %s\nTimestamp: %s\nLog Text: %s", selectedRow[1], selectedRow[2], selectedRow[3])
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	tableView := baseStyle.Render(m.table.View())
	detailView := lipgloss.NewStyle().MarginTop(1).Render(m.detail)
	return tableView + "\n" + detailView
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tally [file or directory]",
	Short: "Analyze a bundle of logfiles for issues",
	Long:  `Analyze a bundle of logfiles for known issues.`,
	Run:   runTally,
	Args:  cobra.MaximumNArgs(1),
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
	width := 120

	columns := []table.Column{
		{Title: "Status", Width: 5},
		{Title: "Source File", Width: 20},
		{Title: "Timestamp", Width: 10},
		{Title: "Log Text", Width: width - 35},
	}

	sort.Slice(logLines, func(i, j int) bool {
		return logLines[i].TimeStamp.Before(logLines[j].TimeStamp)
	})

	var rows []table.Row
	for _, logLine := range logLines {
		rows = append(rows, table.Row{
			logLine.StatusView,
			logLine.FileView,
			logLine.TimeView,
			//logLine.TextView,
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

	m := model{
		table:  t,
		detail: "Select a row to see details here...",
	}
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

	displayLogLines(lines)
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
