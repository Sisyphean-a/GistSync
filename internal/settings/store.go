package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configDirName  = "GistSync"
	configFileName = "settings.json"
)

var ErrEmptySettingsPath = errors.New("settings file path cannot be empty")

type Data struct {
	Token          string `json:"token"`
	MasterPassword string `json:"masterPassword"`
	SyncPath       string `json:"syncPath"`
}

type Store struct {
	filePath string
}

func NewDefaultStore() (*Store, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user config dir: %w", err)
	}
	filePath := filepath.Join(configDir, configDirName, configFileName)
	return NewStore(filePath)
}

func NewStore(filePath string) (*Store, error) {
	if filePath == "" {
		return nil, ErrEmptySettingsPath
	}
	return &Store{filePath: filePath}, nil
}

func (s *Store) Save(data Data) error {
	if err := os.MkdirAll(filepath.Dir(s.filePath), 0o755); err != nil {
		return fmt.Errorf("create settings directory: %w", err)
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	if err = os.WriteFile(s.filePath, raw, 0o600); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}
	return nil
}

func (s *Store) Load() (Data, error) {
	raw, err := os.ReadFile(s.filePath)
	if errors.Is(err, os.ErrNotExist) {
		return Data{}, nil
	}
	if err != nil {
		return Data{}, fmt.Errorf("read settings: %w", err)
	}

	var data Data
	if err = json.Unmarshal(raw, &data); err != nil {
		return Data{}, fmt.Errorf("decode settings: %w", err)
	}
	return data, nil
}
