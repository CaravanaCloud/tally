package cmd

import "time"

type LogLine struct {
	SourceFile string
	fileView string

	Timestamp  time.Time
	timeView string

	Text       string
	textView string
}
