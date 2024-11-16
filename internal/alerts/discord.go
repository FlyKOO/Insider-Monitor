package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/accursedgalaxy/insider-monitor/internal/utils"
)

type DiscordAlerter struct {
    WebhookURL string
    ChannelID  string
}

type discordMessage struct {
    Content   string  `json:"content"`
    Username  string  `json:"username,omitempty"`
    AvatarURL string  `json:"avatar_url,omitempty"`
    Embeds    []embed `json:"embeds,omitempty"`
}

type embed struct {
    Title       string   `json:"title"`
    Description string   `json:"description"`
    Color       int      `json:"color"` // Color code
    Fields      []field  `json:"fields,omitempty"`
}

type field struct {
    Name   string `json:"name"`
    Value  string `json:"value"`
    Inline bool   `json:"inline,omitempty"`
}

func NewDiscordAlerter(webhookURL, channelID string) *DiscordAlerter {
    return &DiscordAlerter{
        WebhookURL: webhookURL,
        ChannelID:  channelID,
    }
}

func (d *DiscordAlerter) SendAlert(alert Alert) error {
    color := 0x7289DA // Default Discord blue
    switch alert.Level {
    case Critical:
        color = 0xFF0000 // Red
    case Warning:
        color = 0xFFA500 // Orange
    }

    // Format the description based on alert type
    var description string
    switch alert.AlertType {
    case "balance_change":
        oldBalance := utils.FormatTokenAmount(alert.Data["old_balance"].(uint64), alert.Data["decimals"].(uint8))
        newBalance := utils.FormatTokenAmount(alert.Data["new_balance"].(uint64), alert.Data["decimals"].(uint8))
        description = fmt.Sprintf("```diff\n%s\n- Old: %s\n+ New: %s\nChange: %+.2f%%```",
            alert.Data["symbol"].(string),
            oldBalance,
            newBalance,
            alert.Data["change_percent"].(float64))

    case "new_token":
        balance := utils.FormatTokenAmount(alert.Data["balance"].(uint64), alert.Data["decimals"].(uint8))
        description = fmt.Sprintf("```ini\n[New Token Added]\nToken: %s\nInitial Balance: %s```",
            alert.Data["symbol"].(string),
            balance)

    case "new_wallet":
        var tokenList strings.Builder
        tokenList.WriteString("```ini\n[Initial Token Balances]\n")
        for symbol, balance := range alert.Data["token_balances"].(map[string]uint64) {
            decimals := alert.Data["token_decimals"].(map[string]uint8)[symbol]
            formatted := utils.FormatTokenAmount(balance, decimals)
            tokenList.WriteString(fmt.Sprintf("%s: %s\n", symbol, formatted))
        }
        tokenList.WriteString("```")
        description = tokenList.String()
    }

    msg := discordMessage{
        Username: "Solana Wallet Monitor",
        Embeds: []embed{{
            Title:       fmt.Sprintf("%s Alert", strings.ToUpper(alert.AlertType)),
            Description: description,
            Color:      color,
            Fields: []field{
                {Name: "Wallet", Value: fmt.Sprintf("`%s`", alert.WalletAddress), Inline: true},
                {Name: "Time", Value: alert.Timestamp.Format("2006-01-02 15:04:05 MST"), Inline: true},
            },
        }},
    }

    payload, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("failed to marshal discord message: %w", err)
    }

    log.Printf("Sending Discord alert: %s", string(payload))

    resp, err := http.Post(d.WebhookURL, "application/json", bytes.NewBuffer(payload))
    if err != nil {
        return fmt.Errorf("failed to send discord message: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("discord API returned error status: %d, body: %s", resp.StatusCode, string(body))
    }

    log.Printf("Successfully sent Discord alert (status: %d)", resp.StatusCode)
    return nil
} 