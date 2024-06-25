package cmd

import "time"

type LogLine struct {
	StatusView string

	FileSource string
	FileView string

	TimeStamp time.Time
	TimeView  string

	Text     string
	TextView string
}
