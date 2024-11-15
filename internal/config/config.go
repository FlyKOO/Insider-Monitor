package config

import (
	"errors"

	"github.com/gagliardetto/solana-go/rpc"
)

type Config struct {
    NetworkURL    string
    Wallets      []string
    ScanInterval string
}

func (c *Config) Validate() error {
    if c.NetworkURL == "" {
        return errors.New("network URL is required")
    }
    if len(c.Wallets) == 0 {
        return errors.New("at least one wallet address is required")
    }
    return nil
}

// Production wallets
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

// Test configuration
func GetTestConfig() *Config {
    return &Config{
        NetworkURL:    rpc.DevNet_RPC,
        Wallets:      []string{"TestWallet1"},
        ScanInterval: "5s",
    }
}