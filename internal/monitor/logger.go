package monitor

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SimpleLogger 同时写入控制台和文件
func LogToFile(dataDir, message string) error {
	logFile := filepath.Join(dataDir, "monitor.log")

	// 使用时间戳格式化日志消息
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)

	// 追加写入文件
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(logMessage)
	return err
}
