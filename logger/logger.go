package logger

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateLogFile create a log file according to the given path and return an opened file.
// Errors are handled by terminating the program.
func CreateLogFile(logPath string, fileName string) *os.File {
	err := os.MkdirAll(logPath, os.ModePerm)
	if err != nil {
		fmt.Printf("Fatal: Cannot mkdir on path %v\n", logPath)
		os.Exit(1)
	}

	logFile := filepath.Join(logPath, fileName)
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Fatal: Cannot open log file %v\n", logFile)
		os.Exit(1)
	}

	return file
}
