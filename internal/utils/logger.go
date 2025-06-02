package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Terminal color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

// Emoji/symbol constants
const (
	InfoSymbol     = "ℹ️"
	SuccessSymbol  = "✅"
	WarningSymbol  = "⚠️"
	ErrorSymbol    = "❌"
	NetworkSymbol  = "🌐"
	WalletSymbol   = "💼"
	ScanSymbol     = "🔍"
	ConfigSymbol   = "⚙️"
	StorageSymbol  = "💾"
	AlertSymbol    = "🔔"
	TimeSymbol     = "⏱️"
	PriceSymbol    = "💰"
	NewTokenSymbol = "🆕"
)

// Logger provides colorful terminal logging
type Logger struct {
	stdLogger *log.Logger
	fileOnly  bool
}

// NewLogger creates a new logger instance
func NewLogger(fileOnly bool) *Logger {
	return &Logger{
		stdLogger: log.New(os.Stdout, "", 0), // No prefix or flags, we'll handle that
		fileOnly:  fileOnly,
	}
}

// formatLog formats a log message with timestamp, emoji, and color
func (l *Logger) formatLog(level, symbol, color, msg string) string {
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	return fmt.Sprintf("%s %s %s%s%s %s",
		timestamp,
		symbol,
		color,
		level,
		ColorReset,
		msg)
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("INFO", InfoSymbol, ColorBlue, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	// Add file logging if needed
	_ = LogToFile("./data", logMsg)
}

// Success logs a success message
func (l *Logger) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("SUCCESS", SuccessSymbol, ColorGreen, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
}

// Warning logs a warning message
func (l *Logger) Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("WARNING", WarningSymbol, ColorYellow, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("ERROR", ErrorSymbol, ColorRed, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
}

// Network logs a network-related message
func (l *Logger) Network(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("NETWORK", NetworkSymbol, ColorCyan, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
}

// Wallet logs a wallet-related message
func (l *Logger) Wallet(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("WALLET", WalletSymbol, ColorPurple, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
}

// Config logs a configuration-related message
func (l *Logger) Config(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("CONFIG", ConfigSymbol, ColorGreen, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
}

// Scan logs a scanning-related message
func (l *Logger) Scan(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("SCAN", ScanSymbol, ColorBlue, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
}

// Storage logs a storage-related message
func (l *Logger) Storage(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("STORAGE", StorageSymbol, ColorCyan, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
}

// Fatal logs a fatal error and exits the program
func (l *Logger) Fatal(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("FATAL", ErrorSymbol, ColorRed, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
	os.Exit(1)
}

// LogToFile writes a log message to a file
func LogToFile(dir string, message string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Use date for log file name
	date := time.Now().Format("2006-01-02")
	logFile := fmt.Sprintf("%s/insider-monitor-%s.log", dir, date)

	// Open log file in append mode
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Write log message with timestamp
	_, err = file.WriteString(message + "\n")
	if err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	return nil
}
