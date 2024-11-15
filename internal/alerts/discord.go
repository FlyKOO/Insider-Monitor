package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
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

    msg := discordMessage{
        Username: "Solana Wallet Monitor",
        Embeds: []embed{{
            Title:       fmt.Sprintf("%s Alert", alert.AlertType),
            Description: alert.Message,
            Color:      color,
            Fields: []field{
                {Name: "Wallet", Value: alert.WalletAddress, Inline: true},
                {Name: "Token", Value: alert.TokenMint, Inline: true},
                {Name: "Time", Value: alert.Timestamp.Format(time.RFC3339), Inline: false},
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