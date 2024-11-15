package monitor

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	bin "github.com/gagliardetto/binary"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
)

const (
	maxRetries = 3
	retryDelay = 2 * time.Second
	// Adjust rate limits based on RPC provider limits
	requestsPerBatch = 5                     // Process in smaller batches
	batchInterval   = 10 * time.Second       // Wait longer between batches
	requestInterval = 500 * time.Millisecond // 2 requests per second
)

type WalletMonitor struct {
	client     *rpc.Client
	wallets    []solana.PublicKey
	networkURL string
	lastRequest time.Time
	mutex       sync.Mutex
}

func NewWalletMonitor(networkURL string, wallets []string) (*WalletMonitor, error) {
	client := rpc.New(networkURL)
	
	// Convert wallet addresses to PublicKeys
	pubKeys := make([]solana.PublicKey, len(wallets))
	for i, addr := range wallets {
		pubKey, err := solana.PublicKeyFromBase58(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid wallet address %s: %v", addr, err)
		}
		pubKeys[i] = pubKey
	}

	return &WalletMonitor{
		client:     client,
		wallets:    pubKeys,
		networkURL: networkURL,
		lastRequest: time.Now(),
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
	Symbol      string    `json:"symbol"`
	Decimals    uint8     `json:"decimals"`
}

// Simplified WalletData
type WalletData struct {
	WalletAddress string                     `json:"wallet_address"`
	TokenAccounts map[string]TokenAccountInfo `json:"token_accounts"` // mint -> info
	LastScanned   time.Time                  `json:"last_scanned"`
}

func (w *WalletMonitor) waitForRateLimit() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	
	elapsed := time.Since(w.lastRequest)
	if elapsed < requestInterval {
		time.Sleep(requestInterval - elapsed)
	}
	w.lastRequest = time.Now()
}

func (w *WalletMonitor) GetWalletData(wallet solana.PublicKey) (*WalletData, error) {
	walletData := &WalletData{
		WalletAddress: wallet.String(),
		TokenAccounts: make(map[string]TokenAccountInfo),
		LastScanned:   time.Now(),
	}

	// Get token accounts with proper encoding
	accounts, err := w.client.GetTokenAccountsByOwner(
		context.Background(),
		wallet,
		&rpc.GetTokenAccountsConfig{
			ProgramId: solana.TokenProgramID.ToPointer(),
		},
		&rpc.GetTokenAccountsOpts{
			Encoding: solana.EncodingBase64Zstd,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get token accounts: %w", err)
	}

	// Process token accounts using binary decoder
	for _, acc := range accounts.Value {
		var tokenAccount token.Account
		err = bin.NewBinDecoder(acc.Account.Data.GetBinary()).Decode(&tokenAccount)
		if err != nil {
			return nil, fmt.Errorf("failed to decode token account: %w", err)
		}

		// Only include accounts with positive balance
		if tokenAccount.Amount > 0 {
			mint := tokenAccount.Mint.String()
			walletData.TokenAccounts[mint] = TokenAccountInfo{
				Balance:     tokenAccount.Amount,
				LastUpdated: time.Now(),
				Symbol:      mint[:8] + "...", // Simplified symbol handling
				Decimals:    uint8(tokenAccount.Amount), // Convert if needed
			}
		}
	}

	log.Printf("Wallet %s: found %d token accounts", wallet.String(), len(walletData.TokenAccounts))
	return walletData, nil
}

// ScanAllWallets scans all monitored wallets and returns their data
func (w *WalletMonitor) ScanAllWallets() (map[string]*WalletData, error) {
	results := make(map[string]*WalletData)
	var errors []error
	
	// Process wallets in batches of 5
	batchSize := 5
	for i := 0; i < len(w.wallets); i += batchSize {
		end := i + batchSize
		if end > len(w.wallets) {
			end = len(w.wallets)
		}
		
		batch := w.wallets[i:end]
		for _, wallet := range batch {
			data, err := w.scanWallet(wallet)
			if err != nil {
				errors = append(errors, fmt.Errorf("wallet %s: %v", wallet, err))
				continue
			}
			if data != nil {
				results[wallet.String()] = data
			}
		}
		
		// Log progress
		log.Printf("Completed batch %d-%d of %d wallets", i+1, end, len(w.wallets))
		
		// Wait between batches to respect rate limits
		if end < len(w.wallets) {
			log.Printf("Waiting 10s before next batch...")
			time.Sleep(10 * time.Second)
		}
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("encountered errors scanning wallets:\n%v", formatErrors(errors))
	}

	return results, nil
}

// Helper function to format multiple errors
func formatErrors(errors []error) string {
	var sb strings.Builder
	for _, err := range errors {
		sb.WriteString("\n")
		sb.WriteString(err.Error())
		sb.WriteString("\n")
	}
	return sb.String()
}

// Update scanWallet to include retries
func (w *WalletMonitor) scanWallet(wallet solana.PublicKey) (*WalletData, error) {
	maxRetries := 3
	retryDelay := 5 * time.Second

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		data, err := w.GetWalletData(wallet)
		if err == nil {
			return data, nil
		}
		
		lastErr = err
		if i < maxRetries-1 {
			// Check if it's a rate limit error
			if rpcErr, ok := err.(*jsonrpc.RPCError); ok && rpcErr.Code == 429 {
				time.Sleep(retryDelay)
				continue
			}
		}
	}
	
	return nil, fmt.Errorf("failed to get wallet data after %d retries: %v", maxRetries, lastErr)
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

func (w *WalletMonitor) getTokenAccountsWithRetry(wallet solana.PublicKey) (*rpc.GetTokenAccountsResult, error) {
	var result *rpc.GetTokenAccountsResult
	var err error

	for i := 0; i < maxRetries; i++ {
		w.waitForRateLimit()
		result, err = w.client.GetTokenAccountsByOwner(
			context.Background(),
			wallet,
			&rpc.GetTokenAccountsConfig{
				ProgramId: solana.TokenProgramID.ToPointer(),
			},
			&rpc.GetTokenAccountsOpts{
				Encoding: solana.EncodingBase64Zstd,
				Commitment: rpc.CommitmentFinalized,
			},
		)
		
		if err == nil {
			// Debug log the successful response
			log.Printf("Successfully retrieved token accounts for %s", wallet.String())
			return result, nil
		}

		// Check if it's a rate limit error
		if strings.Contains(err.Error(), "429") {
			 retryAfter := (i + 1) * 5 // Exponential backoff
			 log.Printf("Rate limited, waiting %d seconds before retry %d/%d", retryAfter, i+1, maxRetries)
			 time.Sleep(time.Duration(retryAfter) * time.Second)
			 continue
		}

		log.Printf("Error getting token accounts (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}
	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}
