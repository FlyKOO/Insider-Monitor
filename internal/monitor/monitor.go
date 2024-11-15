package monitor

import (
	"context"

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
