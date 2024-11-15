package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/accursedgalaxy/insider-monitor/internal/monitor"
)

type Storage struct {
    dataDir string
}

func New(dataDir string) *Storage {
    if err := os.MkdirAll(dataDir, 0755); err != nil {
        log.Printf("warning: failed to create data directory: %v", err)
    }
    return &Storage{dataDir: dataDir}
}

func (s *Storage) SaveWalletData(data map[string]*monitor.WalletData) error {
    // Ensure directory exists before saving
    if err := os.MkdirAll(s.dataDir, 0755); err != nil {
        return fmt.Errorf("failed to create data directory: %w", err)
    }

    path := filepath.Join(s.dataDir, "wallet_data.json")
    file, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal data: %w", err)
    }
    return os.WriteFile(path, file, 0644)
}

func (s *Storage) LoadWalletData() (map[string]*monitor.WalletData, error) {
    path := filepath.Join(s.dataDir, "wallet_data.json")
    data := make(map[string]*monitor.WalletData)
    
    file, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return data, nil
        }
        return nil, err
    }
    
    err = json.Unmarshal(file, &data)
    return data, err
}
