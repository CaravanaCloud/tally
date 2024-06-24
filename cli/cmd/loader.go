package cmd

import (
	"bufio"
	"os"
	"strings"
	"time"
	"path"
)



func loadLinesFromFile(filePath string) ([]LogLine, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var logLines []LogLine
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		logLine, err := parseLogLine(filePath, line)
		if err == nil {
			logLines = append(logLines, logLine)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return logLines, nil
}

func parseLogLine(sourceFile, line string) (LogLine, error) {
	// Try to parse timestamp in RFC3339 format
	var timestamp time.Time
	parts := strings.Fields(line)
	if len(parts) > 0 {
		parsedTime, err := time.Parse(time.RFC3339, parts[0])
		if err == nil {
			timestamp = parsedTime
		} else {
			timestamp = time.Now() // Fallback to current time if parsing fails
		}
	}

	//HELP me HERE: use only file name, without directory ("sample.log" instad of "../samples/sample.log")
	fileView := path.Base(sourceFile)

	return LogLine{
		SourceFile: sourceFile,
		fileView: fileView,
		Timestamp:  timestamp,
		Text:       line,
	}, nil
}
