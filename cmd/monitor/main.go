package main

import (
	"context"
	"log"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)


func main() {
	// Initialize Solana client
	client := rpc.New(rpc.DevNet_RPC) // Using DevNet for testing, change to MainNet for production

	// Test connection
	version, err := client.GetVersion(context.Background())
	if err != nil {
		log.Fatalf("Failed to get Solana version: %v", err)
	}
	log.Printf("Connected to Solana %+v", version)

	// Example wallet address to test
	testWallet := "Gf9XgdmvNHt8fUTFsWAccNbKeyDXsgJyZN8iFJKg5Pbd"
	pubKey := solana.MustPublicKeyFromBase58(testWallet)

	// Get token accounts
	accounts, err := client.GetTokenAccountsByOwner(
		context.Background(),
		pubKey,
		&rpc.GetTokenAccountsConfig{
			ProgramId: solana.TokenProgramID.ToPointer(),
		},
		&rpc.GetTokenAccountsOpts{},
	)
	if err != nil {
		log.Fatalf("Failed to get token accounts: %v", err)
	}


    // Print token accounts
	log.Printf("Found %d token accounts", len(accounts.Value))
    for i, account := range accounts.Value {
        log.Printf("Account %d: %+v", i, account)
    }
}
