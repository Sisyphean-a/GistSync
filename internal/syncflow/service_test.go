package syncflow

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"GistSync/internal/security"
	"GistSync/internal/settings"
)

type fakeCloud struct {
	gistID string
	files  map[string]string
}

func newFakeCloud() *fakeCloud {
	return &fakeCloud{
		gistID: "gist-1",
		files: map[string]string{
			manifestFileName: `{"version":2,"profiles":[],"snapshots":[]}`,
		},
	}
}

func (f *fakeCloud) EnsureManifestGist(context.Context) (string, error) {
	return f.gistID, nil
}

func (f *fakeCloud) UpsertFile(_ context.Context, req UpsertFileRequest) error {
	f.files[req.FileName] = req.Content
	return nil
}

func (f *fakeCloud) GetFileContent(_ context.Context, req FileRequest) (string, error) {
	return f.files[req.FileName], nil
}

func TestService_UploadAndApplySnapshot(t *testing.T) {
	cloud := newFakeCloud()
	service := NewService(cloud)
	sourceFile := filepath.Join(t.TempDir(), "config.txt")
	if err := os.WriteFile(sourceFile, []byte("PM_Secret_Data"), 0o600); err != nil {
		t.Fatalf("write source: %v", err)
	}
	profile := settings.Profile{
		ID:          "profile-1",
		Name:        "Work",
		RestoreMode: restoreOriginal,
		Enabled:     true,
		Items: []settings.ProfileItem{
			{
				ID:                 "item-1",
				SourcePathTemplate: sourceFile,
				RelativePath:       "config.txt",
				Enabled:            true,
			},
		},
	}

	uploadResult, err := service.UploadProfile(context.Background(), UploadProfileRequest{
		Profile: profile, MasterPassword: "123",
	})
	if err != nil {
		t.Fatalf("UploadProfile returned error: %v", err)
	}
	if uploadResult.Uploaded != 1 {
		t.Fatalf("expected one uploaded file, got %d", uploadResult.Uploaded)
	}

	if err = os.WriteFile(sourceFile, []byte("Fake Data"), 0o600); err != nil {
		t.Fatalf("write fake data: %v", err)
	}
	conflicts, err := service.PreviewApplyConflicts(context.Background(), ApplySnapshotRequest{
		ProfileID: profile.ID, SnapshotID: uploadResult.SnapshotID,
	})
	if err != nil {
		t.Fatalf("PreviewApplyConflicts returned error: %v", err)
	}
	if len(conflicts) != 1 {
		t.Fatalf("expected one conflict, got %d", len(conflicts))
	}

	applyResult, err := service.ApplySnapshot(context.Background(), ApplySnapshotRequest{
		ProfileID: profile.ID, SnapshotID: uploadResult.SnapshotID, MasterPassword: "123",
		OverwriteItemIDs: []string{"item-1"},
	})
	if err != nil {
		t.Fatalf("ApplySnapshot returned error: %v", err)
	}
	if applyResult.Applied != 1 {
		t.Fatalf("expected one applied item, got %d", applyResult.Applied)
	}

	raw, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Fatalf("read target file: %v", err)
	}
	if string(raw) != "PM_Secret_Data" {
		t.Fatalf("content mismatch: %s", string(raw))
	}
}

func TestService_ListSnapshots(t *testing.T) {
	cloud := newFakeCloud()
	service := NewService(cloud)
	manifestRaw := `{"version":2,"profiles":[{"id":"profile-1","name":"Work","restoreMode":"original","restoreRoot":"","items":[]}],"snapshots":[{"id":"s1","profileId":"profile-1","createdAt":"2026-03-24T10:00:00Z","items":[]},{"id":"s2","profileId":"profile-1","createdAt":"2026-03-24T11:00:00Z","items":[]}]}`
	cloud.files[manifestFileName] = manifestRaw

	snaps, err := service.ListSnapshots(context.Background(), "profile-1")
	if err != nil {
		t.Fatalf("ListSnapshots returned error: %v", err)
	}
	if len(snaps) != 2 {
		t.Fatalf("snapshot count mismatch: %d", len(snaps))
	}
	if snaps[0].ID != "s2" {
		t.Fatalf("expected latest snapshot first, got %s", snaps[0].ID)
	}
}

func TestService_ListProfilesFromCloud(t *testing.T) {
	cloud := newFakeCloud()
	service := NewService(cloud)
	manifestRaw := `{"version":2,"profiles":[{"id":"profile-1","name":"Work","restoreMode":"original","restoreRoot":"","items":[{"id":"item-1","sourcePathTemplate":"{{HOME}}/.gitconfig","relativePath":".gitconfig","enabled":true}]}],"snapshots":[]}`
	cloud.files[manifestFileName] = manifestRaw

	profiles, err := service.ListProfilesFromCloud(context.Background())
	if err != nil {
		t.Fatalf("ListProfilesFromCloud returned error: %v", err)
	}
	if len(profiles) != 1 || len(profiles[0].Items) != 1 {
		t.Fatalf("unexpected cloud profiles: %#v", profiles)
	}
	if profiles[0].ID != "profile-1" {
		t.Fatalf("profile id mismatch: %s", profiles[0].ID)
	}
}

func TestService_ApplySnapshotRootedMode(t *testing.T) {
	cloud := newFakeCloud()
	service := NewService(cloud)
	rootDir := t.TempDir()
	enc, err := securityEncrypt("hello", "pwd")
	if err != nil {
		t.Fatalf("encrypt helper failed: %v", err)
	}
	manifestData := manifest{
		Version: 2,
		Profiles: []manifestProfile{
			{ID: "p1", Name: "P1", RestoreMode: restoreRooted, Items: []manifestProfileItem{}},
		},
		Snapshots: []manifestSnapshot{
			{
				ID: "s1", ProfileID: "p1", CreatedAt: "2026-03-24T11:00:00Z",
				Items: []manifestSnapshotItem{
					{ItemID: "i1", SourcePathTemplate: "{{HOME}}/.gitconfig", RelativePath: ".gitconfig", BlobFile: "blob1.enc"},
				},
			},
		},
	}
	raw, _ := json.Marshal(manifestData)
	cloud.files[manifestFileName] = string(raw)
	cloud.files["blob1.enc"] = enc

	result, err := service.ApplySnapshot(context.Background(), ApplySnapshotRequest{
		ProfileID: "p1", SnapshotID: "s1", MasterPassword: "pwd",
		RestoreMode: restoreRooted, RestoreRoot: rootDir,
	})
	if err != nil {
		t.Fatalf("ApplySnapshot returned error: %v", err)
	}
	if result.Applied != 1 {
		t.Fatalf("expected apply count 1, got %d", result.Applied)
	}
	targetPath := filepath.Join(rootDir, ".gitconfig")
	out, readErr := os.ReadFile(targetPath)
	if readErr != nil {
		t.Fatalf("read rooted file failed: %v", readErr)
	}
	if string(out) != "hello" {
		t.Fatalf("unexpected rooted content: %q", string(out))
	}
}

func securityEncrypt(data string, password string) (string, error) {
	return security.EncryptString(data, password)
}
