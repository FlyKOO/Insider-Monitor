package config

import (
    "errors"
)

// Config for the WalletMonitor.
type Config struct {
    // NetworkURL is the URL of the Solana network to connect to.
    NetworkURL string
    // Wallets is a list of wallet addresses to monitor.
    Wallets []string
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
    "HUpPyLU8KWisCAr3mzWy2FKT6uuxQ2qGgJQxyTpDoes5",
    "FYGgfgZFeVxnJKF2RS6MKYHBsUpfJdCwumzkPpxWPM4u",
    "GmM5UFm8xu6TnZD7avwYcQ1zq25hD5yvHfYyAksHu9vB",
    "CWvdyvKHEu8Z6QqGraJT3sLPyp9bJfFhoXcxUYRKC8ou",
}
