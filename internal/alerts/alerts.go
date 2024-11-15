package alerts

import (
	"fmt"
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
    Data          map[string]interface{} // Additional data
}

type Alerter interface {
    SendAlert(alert Alert) error
}

// ConsoleAlerter implements basic console logging
type ConsoleAlerter struct{}

func (a *ConsoleAlerter) SendAlert(alert Alert) error {
    fmt.Printf("[%s] %s: %s\n", alert.Level, alert.Timestamp.Format(time.RFC3339), alert.Message)
    return nil
} 