package settings

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStore_SaveLoadRoundTripV2(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "settings.json")
	store, err := NewStore(filePath)
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	input := Data{
		Token:           "ABC",
		MasterPassword:  "123",
		ActiveProfileID: "profile-1",
		Profiles: []Profile{
			{
				ID:          "profile-1",
				Name:        "Work",
				RestoreMode: restoreOriginal,
				Enabled:     true,
				Items: []ProfileItem{
					{
						ID:                 "item-1",
						SourcePathTemplate: "{{HOME}}/.gitconfig",
						RelativePath:       ".gitconfig",
						Enabled:            true,
					},
				},
			},
		},
	}
	if err = store.Save(input); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if got.Token != input.Token || got.MasterPassword != input.MasterPassword {
		t.Fatalf("credentials mismatch: got %#v want %#v", got, input)
	}
	if len(got.Profiles) != 1 || len(got.Profiles[0].Items) != 1 {
		t.Fatalf("profile data mismatch: got %#v", got)
	}
}

func TestStore_LoadMissingFile(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "missing.json")
	store, err := NewStore(filePath)
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if got.Token != "" || got.MasterPassword != "" || got.ActiveProfileID != "" || len(got.Profiles) != 0 {
		t.Fatalf("expected zero value data for missing file, got %#v", got)
	}
}

func TestStore_MigrateLegacySyncPath(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "settings.json")
	legacyJSON := `{"token":"ABC","masterPassword":"123","syncPath":"{{HOME}}/.ssh/config"}`
	if err := os.WriteFile(filePath, []byte(legacyJSON), 0o600); err != nil {
		t.Fatalf("write legacy file: %v", err)
	}

	store, err := NewStore(filePath)
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(got.Profiles) != 1 || len(got.Profiles[0].Items) != 1 {
		t.Fatalf("legacy migration failed, got %#v", got)
	}
	if strings.TrimSpace(got.ActiveProfileID) == "" {
		t.Fatalf("expected active profile after migration")
	}
	if !got.CloudBootstrapDone {
		t.Fatalf("expected cloud bootstrap done after migration")
	}
}

func TestStore_LoadMissingFileBootstrapFlag(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "missing.json")
	store, err := NewStore(filePath)
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if got.CloudBootstrapDone {
		t.Fatalf("expected cloud bootstrap flag to be false for missing file")
	}
}
