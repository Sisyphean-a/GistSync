package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	configDirName   = "GistSync"
	configFileName  = "settings.json"
	restoreOriginal = "original"
	restoreRooted   = "rooted"
)

var ErrEmptySettingsPath = errors.New("settings file path cannot be empty")

type ProfileItem struct {
	ID                 string `json:"id"`
	SourcePathTemplate string `json:"sourcePathTemplate"`
	RelativePath       string `json:"relativePath"`
	Enabled            bool   `json:"enabled"`
}

type Profile struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	RestoreMode string        `json:"restoreMode"`
	RestoreRoot string        `json:"restoreRoot"`
	Enabled     bool          `json:"enabled"`
	Items       []ProfileItem `json:"items"`
}

type Data struct {
	Token              string    `json:"token"`
	MasterPassword     string    `json:"masterPassword"`
	ActiveProfileID    string    `json:"activeProfileId"`
	Profiles           []Profile `json:"profiles"`
	CloudBootstrapDone bool      `json:"cloudBootstrapDone,omitempty"`
	SyncPath           string    `json:"syncPath,omitempty"`
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
	if strings.TrimSpace(filePath) == "" {
		return nil, ErrEmptySettingsPath
	}
	return &Store{filePath: filePath}, nil
}

func (s *Store) Save(data Data) error {
	data = applyDefaults(data)
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
	data = migrateLegacy(data)
	return applyDefaults(data), nil
}

func applyDefaults(data Data) Data {
	if len(data.Profiles) == 0 {
		data.ActiveProfileID = ""
		return data
	}
	data.CloudBootstrapDone = true
	for i := range data.Profiles {
		profile := &data.Profiles[i]
		if profile.ID == "" {
			profile.ID = generateID("profile")
		}
		if strings.TrimSpace(profile.Name) == "" {
			profile.Name = autoProfileName()
		}
		if profile.RestoreMode == "" {
			profile.RestoreMode = restoreOriginal
		}
		for j := range profile.Items {
			item := &profile.Items[j]
			if item.ID == "" {
				item.ID = generateID("item")
			}
		}
	}

	activeExists := false
	for _, profile := range data.Profiles {
		if profile.ID == data.ActiveProfileID {
			activeExists = true
			break
		}
	}
	if data.ActiveProfileID == "" || !activeExists {
		data.ActiveProfileID = data.Profiles[0].ID
	}
	return data
}

func migrateLegacy(data Data) Data {
	if strings.TrimSpace(data.SyncPath) == "" || len(data.Profiles) > 0 {
		return data
	}
	profileID := generateID("profile")
	itemID := generateID("item")
	data.Profiles = []Profile{
		{
			ID:          profileID,
			Name:        "迁移配置",
			RestoreMode: restoreOriginal,
			Enabled:     true,
			Items: []ProfileItem{
				{
					ID:                 itemID,
					SourcePathTemplate: data.SyncPath,
					RelativePath:       buildRelativePath(data.SyncPath),
					Enabled:            true,
				},
			},
		},
	}
	data.ActiveProfileID = profileID
	data.SyncPath = ""
	return data
}

func buildRelativePath(sourcePath string) string {
	path := strings.ReplaceAll(sourcePath, "\\", "/")
	path = strings.TrimPrefix(path, "{{HOME}}/")
	path = strings.TrimPrefix(path, "/")
	path = strings.ReplaceAll(path, ":", "")
	return path
}

func autoProfileName() string {
	return "配置-" + time.Now().Format("20060102-150405")
}

func generateID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func IsValidRestoreMode(mode string) bool {
	return mode == restoreOriginal || mode == restoreRooted
}
