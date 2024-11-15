package config

import (
	"errors"

	"github.com/gagliardetto/solana-go/rpc"
)

// Config for the WalletMonitor.
type Config struct {
    // NetworkURL is the URL of the Solana network to connect to.
    NetworkURL string
    // Wallets is a list of wallet addresses to monitor.
    Wallets []string
    // ScanInterval is the interval between scans.
    ScanInterval string
}

// Validate the Config.
func (c *Config) Validate() error {
    if c.NetworkURL == "" {
        return errors.New("NetworkURL is required")
    }
    if len(c.Wallets) == 0 {
        return errors.New("At least one wallet address is required")
    }
    return nil
}

// List of wallet addresses to monitor.
var Wallets = []string{
    "Gf9XgdmvNHt8fUTFsWAccNbKeyDXsgJyZN8iFJKg5Pbd",
    // "HUpPyLU8KWisCAr3mzWy2FKT6uuxQ2qGgJQxyTpDoes5",
    // "FYGgfgZFeVxnJKF2RS6MKYHBsUpfJdCwumzkPpxWPM4u",
    // "GmM5UFm8xu6TnZD7avwYcQ1zq25hD5yvHfYyAksHu9vB",
    // "CWvdyvKHEu8Z6QqGraJT3sLPyp9bJfFhoXcxUYRKC8ou",
}

var TestWallets = []string{
    "55kBY9yxqQzj2zxZqRkqENYq6R8PkXmn5GKyQN9YeVFr", // Known test wallet
    "DWuopnuSqYdBhCXqxfqjqzPGibnhkj6SQqFvgC4jkvjF", // Another test wallet
}

func GetTestConfig() *Config {
    return &Config{
        NetworkURL: rpc.DevNet_RPC,
        Wallets:    TestWallets,
        ScanInterval: "5s",
    }
}
