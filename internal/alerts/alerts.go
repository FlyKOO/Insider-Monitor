package alerts

import (
	"time"
)

type AlertLevel string

const (
	Info     AlertLevel = "INFO"
	Warning  AlertLevel = "WARNING"
	Critical AlertLevel = "CRITICAL"
)

type Alert struct {
	Timestamp     time.Time
	WalletAddress string
	TokenMint     string
	AlertType     string
	Message       string
	Level         AlertLevel
	Data          map[string]interface{} // 用于格式化的附加数据
}

type Alerter interface {
	SendAlert(alert Alert) error
}

// ConsoleAlerter 的实现已移动至 console.go
