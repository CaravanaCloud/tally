package cmd

import (
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// GetFileName extracts the file name from a given file path
func GetFileName(path string) string {
	// Split the path by the path separator
	parts := strings.Split(path, "/")

	// Return the last part, which is the file name
	return parts[len(parts)-1]
}


// ExtractTimePart extracts the time part from an ISO 8601 timestamp
func ExtractTimePart(t time.Time) string {
	// Format the time part to a string in "HH:MM:SS" format
	return t.Format("15:04:05")
}


var (
	timeHighlightStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	ipHighlightStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("105"))
	httpMethodStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	errorHighlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("160"))
)

// isValidTimestamp checks if a string matches the specified timestamp format.
func isValidTimestamp(timestamp string) bool {
	re := regexp.MustCompile(`^\d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2} \+\d{4}$`)
	return re.MatchString(timestamp)
}

// HighlightTimestamp replaces the first matching timestamp with only the time part highlighted.
func HighlightTimestamp(line string) string {
	re := regexp.MustCompile(`(\d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2} \+\d{4})`)
	return re.ReplaceAllStringFunc(line, func(match string) string {
		if isValidTimestamp(match) {
			timePart := match[strings.Index(match, ":")+1 : strings.Index(match, " ")]
			return timeHighlightStyle.Render(timePart)
		}
		return match
	})
}

// ObfuscateIP replaces any IP address with the last part followed by "..."
func ObfuscateIP(line string) string {
	re := regexp.MustCompile(`(\d{1,3}\.){3}\d{1,3}`)
	return re.ReplaceAllStringFunc(line, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) > 0 {
			lastPart := match[strings.LastIndex(match, ".")+1:]
			return ipHighlightStyle.Render("..."+lastPart)
		}
		return match
	})
}

// HighlightHTTPMethods highlights HTTP method names.
func HighlightHTTPMethods(line string) string {
	re := regexp.MustCompile(`\b(GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD)\b`)
	return re.ReplaceAllStringFunc(line, func(match string) string {
		return httpMethodStyle.Render(match)
	})
}

// HighlightHTTPStatus highlights HTTP status codes (400-499, 500-599).
func HighlightHTTPStatus(line string) string {
	re := regexp.MustCompile(`\b(4\d{2}|5\d{2})\b`)
	return re.ReplaceAllStringFunc(line, func(match string) string {
		return errorHighlightStyle.Render(match)
	})
}

// SummarizeLine applies all transformations to the log line.
func SummarizeLine(orig string) string {
	line := orig
	line = HighlightTimestamp(line)
	line = ObfuscateIP(line)
	line = HighlightHTTPMethods(line)
	line = HighlightHTTPStatus(line)
	return line
}

// StatusOf returns "?" if the line contains "error" or an HTTP status code, otherwise " ".
func StatusOf(line string) string {
	// Regular expression to match "error" or HTTP status codes (400-599)
	re := regexp.MustCompile(`\berror\b|\b(4\d{2}|5\d{2})\b`)
	if re.MatchString(strings.ToLower(line)) {
		return "?"
	}
	return " "
}


