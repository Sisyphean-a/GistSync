package appsvc

import (
	"context"
	"errors"
	"strings"

	"GistSync/internal/gistapi"
	"GistSync/internal/profileutil"
	"GistSync/internal/settings"
	"GistSync/internal/syncflow"
)

var ErrOverwriteRequired = errors.New("OVERWRITE_REQUIRED")

type SettingsStore interface {
	Load() (settings.Data, error)
	Save(data settings.Data) error
}

type SyncService interface {
	ListProfilesFromCloud(ctx context.Context) ([]settings.Profile, error)
	UploadProfile(ctx context.Context, req syncflow.UploadProfileRequest) (syncflow.UploadProfileResult, error)
	ListSnapshots(ctx context.Context, profileID string) ([]syncflow.SnapshotMeta, error)
	PreviewApplyConflicts(ctx context.Context, req syncflow.ApplySnapshotRequest) ([]syncflow.ApplyConflict, error)
	ApplySnapshot(ctx context.Context, req syncflow.ApplySnapshotRequest) (syncflow.ApplySnapshotResult, error)
}

type SyncFactory func(token string) (SyncService, error)

type Service struct {
	store      SettingsStore
	buildSync  SyncFactory
	generateID func(prefix string) string
}

func NewService(store SettingsStore) *Service {
	return &Service{
		store:      store,
		buildSync:  defaultSyncFactory,
		generateID: profileutil.GenerateID,
	}
}

func defaultSyncFactory(token string) (SyncService, error) {
	client, err := gistapi.NewClient(gistapi.ClientOptions{Token: token})
	if err != nil {
		return nil, err
	}
	return syncflow.NewService(syncflow.NewGistGateway(client)), nil
}

func shouldBootstrap(data settings.Data) bool {
	return len(data.Profiles) == 0 && !empty(data.Token) && !data.CloudBootstrapDone
}

func empty(value string) bool {
	return strings.TrimSpace(value) == ""
}

func findProfile(data settings.Data, profileID string) (*settings.Profile, bool) {
	for i := range data.Profiles {
		if data.Profiles[i].ID == profileID {
			return &data.Profiles[i], true
		}
	}
	return nil, false
}

func resolveProfileID(data settings.Data, requestedID string) string {
	if !empty(requestedID) {
		if _, ok := findProfile(data, requestedID); ok {
			return requestedID
		}
	}
	if !empty(data.ActiveProfileID) {
		if _, ok := findProfile(data, data.ActiveProfileID); ok {
			return data.ActiveProfileID
		}
	}
	if len(data.Profiles) == 0 {
		return ""
	}
	return data.Profiles[0].ID
}
