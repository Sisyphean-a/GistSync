package appsvc

import (
	"context"
	"errors"
	"testing"

	"GistSync/internal/settings"
	"GistSync/internal/syncflow"
)

type fakeStore struct {
	data settings.Data
}

func (s *fakeStore) Load() (settings.Data, error) {
	return s.data, nil
}

func (s *fakeStore) Save(data settings.Data) error {
	s.data = data
	return nil
}

type fakeSync struct {
	preview []syncflow.ApplyConflict
	apply   syncflow.ApplySnapshotResult
}

func (f *fakeSync) ListProfilesFromCloud(context.Context) ([]settings.Profile, error) {
	return nil, nil
}

func (f *fakeSync) UploadProfile(context.Context, syncflow.UploadProfileRequest) (syncflow.UploadProfileResult, error) {
	return syncflow.UploadProfileResult{}, nil
}

func (f *fakeSync) ListSnapshots(context.Context, string) ([]syncflow.SnapshotMeta, error) {
	return nil, nil
}

func (f *fakeSync) PreviewApplyConflicts(context.Context, syncflow.ApplySnapshotRequest) ([]syncflow.ApplyConflict, error) {
	return f.preview, nil
}

func (f *fakeSync) ApplySnapshot(context.Context, syncflow.ApplySnapshotRequest) (syncflow.ApplySnapshotResult, error) {
	return f.apply, nil
}

func TestService_AddFilesToProfile_NormalizesRelativePath(t *testing.T) {
	store := &fakeStore{data: settings.Data{Profiles: []settings.Profile{{ID: "p1", Items: []settings.ProfileItem{}}}}}
	svc := NewService(store)
	svc.generateID = func(prefix string) string { return prefix + "-id" }

	err := svc.AddFilesToProfile(context.Background(), "p1", []string{`C:\\Users\\me\\.gitconfig`})
	if err != nil {
		t.Fatalf("AddFilesToProfile returned error: %v", err)
	}
	item := store.data.Profiles[0].Items[0]
	if item.RelativePath != "Users/me/.gitconfig" {
		t.Fatalf("relative path mismatch: %q", item.RelativePath)
	}
}

func TestService_DownloadSync_RequiresOverwrite(t *testing.T) {
	store := &fakeStore{data: settings.Data{Token: "t1", ActiveProfileID: "p1", Profiles: []settings.Profile{{ID: "p1"}}}}
	syncSvc := &fakeSync{preview: []syncflow.ApplyConflict{{ItemID: "i1", TargetPath: "/tmp/x"}}}
	svc := NewService(store)
	svc.buildSync = func(token string) (SyncService, error) {
		if token != "t1" {
			t.Fatalf("unexpected token %q", token)
		}
		return syncSvc, nil
	}

	_, err := svc.DownloadSync(context.Background(), false)
	if !errors.Is(err, ErrOverwriteRequired) {
		t.Fatalf("expected ErrOverwriteRequired, got %v", err)
	}
}
