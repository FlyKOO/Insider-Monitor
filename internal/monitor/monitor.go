package monitor

import (
	"context"
	"encoding/binary"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type WalletMonitor struct {
	client     *rpc.Client
	wallets    []solana.PublicKey
	networkURL string
}

func NewWalletMonitor(networkURL string, wallets []string) (*WalletMonitor, error) {
	client := rpc.New(networkURL)

	// Convert wallet addresses to PublicKeys
	walletKeys := make([]solana.PublicKey, len(wallets))
	for i, wallet := range wallets {
		walletKeys[i] = solana.MustPublicKeyFromBase58(wallet)
	}

	return &WalletMonitor{
		client:     client,
		wallets:    walletKeys,
		networkURL: networkURL,
	}, nil
}

func (w *WalletMonitor) GetTokenAccounts(wallet solana.PublicKey) (*rpc.GetTokenAccountsResult, error) {
	return w.client.GetTokenAccountsByOwner(
		context.Background(),
		wallet,
		&rpc.GetTokenAccountsConfig{
			ProgramId: solana.TokenProgramID.ToPointer(),
		},
		&rpc.GetTokenAccountsOpts{},
	)
}

// Simplified TokenAccountInfo
type TokenAccountInfo struct {
	Balance     uint64    `json:"balance"`
	LastUpdated time.Time `json:"last_updated"`
}

// Simplified WalletData
type WalletData struct {
	WalletAddress string                     `json:"wallet_address"`
	TokenAccounts map[string]TokenAccountInfo `json:"token_accounts"` // mint -> info
	LastScanned   time.Time                  `json:"last_scanned"`
}

func (w *WalletMonitor) GetWalletData(wallet solana.PublicKey) (*WalletData, error) {
	accounts, err := w.client.GetTokenAccountsByOwner(
		context.Background(),
		wallet,
		&rpc.GetTokenAccountsConfig{
			ProgramId: solana.TokenProgramID.ToPointer(),
		},
		&rpc.GetTokenAccountsOpts{},
	)
	if err != nil {
		return nil, err
	}

	walletData := &WalletData{
		WalletAddress: wallet.String(),
		TokenAccounts: make(map[string]TokenAccountInfo),
		LastScanned:   time.Now(),
	}

	for _, acc := range accounts.Value {
		data := acc.Account.Data.GetBinary()
		if len(data) < 165 {
			continue
		}

		mint := solana.PublicKey(data[0:32])
		balance := binary.LittleEndian.Uint64(data[64:72])

		if balance > 0 {
			walletData.TokenAccounts[mint.String()] = TokenAccountInfo{
				Balance:     balance,
				LastUpdated: time.Now(),
			}
		}
	}

	return walletData, nil
}

// ScanAllWallets scans all monitored wallets and returns their data
func (w *WalletMonitor) ScanAllWallets() (map[string]*WalletData, error) {
	results := make(map[string]*WalletData)
	
	for _, wallet := range w.wallets {
		data, err := w.GetWalletData(wallet)
		if err != nil {
			continue // Skip failed wallets but continue scanning others
		}
		results[wallet.String()] = data
	}

	return results, nil
}
