package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// splitLines splits the content into individual lines.
func splitLines(content string) []string {
	return strings.Split(strings.TrimSpace(content), "\n")
}

// List of common time format strings
var timeFormats = []string{
	"02/Jan/2006:15:04:05 -0700", // Common HTTP access log format
	"2006-01-02T15:04:05Z07:00",  // ISO 8601 format
	"Mon Jan 2 15:04:05 2006",    // Unix date format
}

// Regular expression to extract the timestamp part from the log line
var timeRegex = regexp.MustCompile(`\[(\d{2}/\w{3}/\d{4}:\d{2}:\d{2}:\d{2} \+\d{4})\]|\b(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z?\d{2}:\d{2})\b|\b(\w{3} \w{3} \d{2} \d{2}:\d{2}:\d{2} \d{4})\b`)

// parseAny tries to parse the timestamp using multiple formats
func parseAny(timestamp string) (time.Time, error) {
	for _, format := range timeFormats {
		parsedTime, err := time.Parse(format, timestamp)
		if err == nil {
			return parsedTime, nil
		}
	}
	return time.Time{}, fmt.Errorf("no matching format found")
}

// parseTime extracts and parses the timestamp from a log line
func parseTime(text string) time.Time {
	// Find the first match
	matches := timeRegex.FindStringSubmatch(text)
	if len(matches) < 2 {
		fmt.Println("Error: No timestamp found in log line")
		return time.Time{} // Return zero value of time.Time on error
	}

	// Extracted timestamp
	var timestamp string
	for _, match := range matches[1:] {
		if match != "" {
			timestamp = match
			break
		}
	}

	// Parse the extracted timestamp using parseAny
	parsedTime, err := parseAny(timestamp)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return time.Time{} // Return zero value of time.Time on error
	}

	return parsedTime
}

func parseTemperature(logLine string) int8 {
	// Regular expression to match HTTP method names and error status codes (40X or 50X)
	re := regexp.MustCompile(`"(GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD) [^"]* HTTP/1\.[01]" (4\d{2}|5\d{2})`)

	// Find the first match
	matches := re.FindStringSubmatch(logLine)
	if len(matches) > 0 {
		statusCode, err := strconv.Atoi(matches[2])
		if err == nil && (statusCode >= 400 && statusCode < 600) {
			return 100
		}
	}

	return 0
}

// loadFilesInDirectory reads all log files from the specified directory recursively.
func loadFilesInDirectory(dirPath string) error {
	mu.Lock() // Lock to prevent concurrent access
	defer mu.Unlock()

	lines = nil

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
			splitLines := splitLines(string(content))
			for _, line := range splitLines {
				lineTime := parseTime(line)
				lineTemp := parseTemperature(line)
				logline := LogLine{
					text:        line,
					time:        lineTime,
					temperature: lineTemp,
				}
				lines = append(lines, logline)
			}
			log.Printf("Loaded file: %s (%d lines)\n", path, len(splitLines)) // Debug output
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error gathering log files: %w", err)
	}

	if len(lines) == 0 {
		fmt.Println("No log files found in directory.")
	}
	scrollOffset = len(lines) - getTerminalHeight() + 1
	if scrollOffset < 0 {
		scrollOffset = 0
	}
	selectedLine = len(lines) - 1
	// sort lines by time
	sort.Slice(lines, func(i, j int) bool {
		return lines[i].time.Before(lines[j].time)
	})
	render()
	return nil
}
