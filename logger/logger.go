package logger

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateLogFile create a log file according to the given path and return an opened file.
// Errors are handled by terminating the program.
func CreateLogFile(logPath string) *os.File {
	err := os.MkdirAll(logPath, os.ModePerm)
	if err != nil {
		fmt.Printf("Fatal: Cannot mkdir on path %v\n", logPath)
		os.Exit(1)
	}

	logFile := filepath.Join(logPath, "shortener.log")
	file, err := os.Open(logFile)
	if os.IsNotExist(err) {
		_, e := os.Create(filepath.Join(logPath, "shortener.log"))
		if e != nil {
			fmt.Printf("Fatal: Cannot create log file %v\n", logFile)
			os.Exit(1)
		}
	} else if err != nil {
		fmt.Printf("Fatal: Cannot open log file %v\n", logFile)
		os.Exit(1)
	}

	return file
}
