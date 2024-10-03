package main

import (
	"log"

	"github.com/CaravanaCloud/tally/cmd" // Replace "your_project" with your actual module name
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
