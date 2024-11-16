package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

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
    
    // Create storage directory if it doesn't exist
    if err := os.MkdirAll(s.dataDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create data directory: %w", err)
    }

    // If file doesn't exist, create empty data
    if _, err := os.Stat(path); os.IsNotExist(err) {
        emptyData := make(map[string]*monitor.WalletData)
        if err := s.SaveWalletData(emptyData); err != nil {
            return nil, fmt.Errorf("failed to create initial data file: %w", err)
        }
        return emptyData, nil
    }

    data := make(map[string]*monitor.WalletData)
    file, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return data, nil
        }
        return nil, err
    }
    
    err = json.Unmarshal(file, &data)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal data: %w", err)
    }
    
    return data, nil
}

func (s *Storage) IsDataValid() bool {
    data, err := s.LoadWalletData()
    if err != nil {
        return false
    }
    return len(data) > 0
}

func (s *Storage) BackupCurrentData() error {
    currentData, err := s.LoadWalletData()
    if err != nil {
        return err
    }

    backupPath := filepath.Join(s.dataDir, fmt.Sprintf("wallet_data_backup_%d.json", time.Now().Unix()))
    file, err := json.MarshalIndent(currentData, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(backupPath, file, 0644)
}
