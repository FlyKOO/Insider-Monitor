package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

// 终端颜色代码
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

// 表情与符号常量
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

// Logger 提供带颜色的终端日志
type Logger struct {
	stdLogger *log.Logger
	fileOnly  bool
}

// NewLogger 创建一个新的日志实例
func NewLogger(fileOnly bool) *Logger {
	return &Logger{
		stdLogger: log.New(os.Stdout, "", 0), // 无前缀和标志，由我们自行处理
		fileOnly:  fileOnly,
	}
}

// formatLog 按时间戳、表情和颜色格式化日志消息
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

// Info 记录一条信息级别日志
func (l *Logger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("INFO", InfoSymbol, ColorBlue, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	// 如有需要添加文件日志
	_ = LogToFile("./data", logMsg)
}

// Success 记录一条成功日志
func (l *Logger) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("SUCCESS", SuccessSymbol, ColorGreen, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
}

// Warning 记录一条警告日志
func (l *Logger) Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("WARNING", WarningSymbol, ColorYellow, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
}

// Error 记录一条错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("ERROR", ErrorSymbol, ColorRed, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
}

// Network 记录一条网络相关日志
func (l *Logger) Network(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("NETWORK", NetworkSymbol, ColorCyan, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
}

// Wallet 记录一条钱包相关日志
func (l *Logger) Wallet(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("WALLET", WalletSymbol, ColorPurple, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
}

// Config 记录一条配置相关日志
func (l *Logger) Config(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("CONFIG", ConfigSymbol, ColorGreen, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
}

// Scan 记录一条扫描相关日志
func (l *Logger) Scan(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("SCAN", ScanSymbol, ColorBlue, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
}

// Storage 记录一条存储相关日志
func (l *Logger) Storage(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("STORAGE", StorageSymbol, ColorCyan, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
}

// Fatal 记录一条致命错误并退出程序
func (l *Logger) Fatal(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logMsg := l.formatLog("FATAL", ErrorSymbol, ColorRed, msg)
	if !l.fileOnly {
		l.stdLogger.Println(logMsg)
	}
	_ = LogToFile("./data", logMsg)
	os.Exit(1)
}

// LogToFile 将日志信息写入文件
func LogToFile(dir string, message string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// 使用日期作为日志文件名
	date := time.Now().Format("2006-01-02")
	logFile := fmt.Sprintf("%s/insider-monitor-%s.log", dir, date)

	// 以追加模式打开日志文件
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// 写入带时间戳的日志信息
	_, err = file.WriteString(message + "\n")
	if err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	return nil
}
