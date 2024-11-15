package monitor

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SimpleLogger writes to both console and file
func LogToFile(dataDir, message string) error {
	logFile := filepath.Join(dataDir, "monitor.log")
	
	// Format the log message with timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)
	
	// Append to file
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	
	_, err = f.WriteString(logMessage)
	return err
} 