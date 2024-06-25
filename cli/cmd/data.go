package cmd

import "time"

type LogLine struct {
	FileSource string
	FileView string

	TimeStamp time.Time
	TimeView  string

	Text     string
	TextView string
}
