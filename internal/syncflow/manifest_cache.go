package syncflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ManifestCache interface {
	Load(gistID string, maxAge time.Duration) (manifest, bool, error)
	Save(gistID string, data manifest) error
}

type fileManifestCache struct {
	rootDir string
}

type cachedManifest struct {
	SavedAt string   `json:"savedAt"`
	Data    manifest `json:"data"`
}

func newFileManifestCache() ManifestCache {
	rootDir, err := resolveManifestCacheDir()
	if err != nil {
		return &noopManifestCache{}
	}
	return &fileManifestCache{rootDir: rootDir}
}

func resolveManifestCacheDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "GistSync", "manifest-cache"), nil
}

func (c *fileManifestCache) Load(gistID string, maxAge time.Duration) (manifest, bool, error) {
	path := c.cachePath(gistID)
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return manifest{}, false, nil
	}
	if err != nil {
		return manifest{}, false, fmt.Errorf("read manifest cache: %w", err)
	}
	var record cachedManifest
	if err = json.Unmarshal(raw, &record); err != nil {
		return manifest{}, false, fmt.Errorf("decode manifest cache: %w", err)
	}
	savedAt, err := time.Parse(time.RFC3339, record.SavedAt)
	if err != nil {
		return manifest{}, false, fmt.Errorf("parse manifest cache time: %w", err)
	}
	if time.Since(savedAt) > maxAge {
		return manifest{}, false, nil
	}
	return record.Data, true, nil
}

func (c *fileManifestCache) Save(gistID string, data manifest) error {
	if err := os.MkdirAll(c.rootDir, 0o755); err != nil {
		return fmt.Errorf("create manifest cache dir: %w", err)
	}
	record := cachedManifest{
		SavedAt: time.Now().UTC().Format(time.RFC3339),
		Data:    data,
	}
	raw, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encode manifest cache: %w", err)
	}
	if err = os.WriteFile(c.cachePath(gistID), raw, 0o600); err != nil {
		return fmt.Errorf("write manifest cache: %w", err)
	}
	return nil
}

func (c *fileManifestCache) cachePath(gistID string) string {
	return filepath.Join(c.rootDir, gistID+".json")
}

type noopManifestCache struct{}

func (c *noopManifestCache) Load(string, time.Duration) (manifest, bool, error) {
	return manifest{}, false, nil
}

func (c *noopManifestCache) Save(string, manifest) error {
	return nil
}
