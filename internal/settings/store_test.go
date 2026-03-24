package settings

import (
	"path/filepath"
	"testing"
)

func TestStore_SaveLoadRoundTrip(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "settings.json")
	store, err := NewStore(filePath)
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	want := Data{
		Token:          "ABC",
		MasterPassword: "123",
		SyncPath:       "{{HOME}}/Desktop/test-config.txt",
	}
	if err = store.Save(want); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if got != want {
		t.Fatalf("settings mismatch: got %#v want %#v", got, want)
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
	if got != (Data{}) {
		t.Fatalf("expected zero value data for missing file, got %#v", got)
	}
}
