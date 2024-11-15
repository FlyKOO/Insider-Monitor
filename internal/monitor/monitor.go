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

// Change represents a detected change in a wallet
type Change struct {
	WalletAddress string
	TokenMint     string
	ChangeType    string // "new_wallet", "new_token", "balance_change"
	OldBalance    uint64
	NewBalance    uint64
}

// DetectChanges compares old and new wallet data to find changes
func DetectChanges(old, new map[string]*WalletData) []Change {
	var changes []Change
	
	for wallet, newData := range new {
		oldData, existed := old[wallet]
		
		// Check for new wallet
		if !existed {
			for mint, token := range newData.TokenAccounts {
				changes = append(changes, Change{
					WalletAddress: wallet,
					TokenMint:     mint,
					ChangeType:    "new_wallet",
					NewBalance:    token.Balance,
				})
			}
			continue
		}
		
		// Check for token changes
		for mint, newToken := range newData.TokenAccounts {
			oldToken, hadToken := oldData.TokenAccounts[mint]
			if !hadToken {
				changes = append(changes, Change{
					WalletAddress: wallet,
					TokenMint:     mint,
					ChangeType:    "new_token",
					NewBalance:    newToken.Balance,
				})
			} else if oldToken.Balance != newToken.Balance {
				changes = append(changes, Change{
					WalletAddress: wallet,
					TokenMint:     mint,
					ChangeType:    "balance_change",
					OldBalance:    oldToken.Balance,
					NewBalance:    newToken.Balance,
				})
			}
		}
	}
	
	return changes
}
