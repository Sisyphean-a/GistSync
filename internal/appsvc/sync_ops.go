package appsvc

import (
	"context"
	"errors"
	"fmt"

	"GistSync/internal/settings"
	"GistSync/internal/syncflow"
)

func (s *Service) UploadProfile(ctx context.Context, profileID string, selectedItemIDs []string) (syncflow.UploadProfileResult, error) {
	data, profile, syncService, err := s.loadSyncContext(profileID)
	if err != nil {
		return syncflow.UploadProfileResult{}, err
	}
	return syncService.UploadProfile(ctx, syncflow.UploadProfileRequest{
		Profile: profile, MasterPassword: data.MasterPassword, SelectedItemIDs: selectedItemIDs,
	})
}

func (s *Service) ListSnapshots(ctx context.Context, profileID string) ([]syncflow.SnapshotMeta, error) {
	data, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	resolvedID := resolveProfileID(data, profileID)
	if empty(resolvedID) {
		return []syncflow.SnapshotMeta{}, nil
	}
	syncService, err := s.buildSync(data.Token)
	if err != nil {
		return nil, err
	}
	snapshots, err := syncService.ListSnapshots(ctx, resolvedID)
	if errors.Is(err, syncflow.ErrProfileNotFound) {
		return []syncflow.SnapshotMeta{}, nil
	}
	return snapshots, err
}

func (s *Service) PreviewApplyConflicts(ctx context.Context, req syncflow.ApplySnapshotRequest) ([]syncflow.ApplyConflict, error) {
	data, _, syncService, err := s.loadSyncContext(req.ProfileID)
	if err != nil {
		return nil, err
	}
	if empty(req.MasterPassword) {
		req.MasterPassword = data.MasterPassword
	}
	return syncService.PreviewApplyConflicts(ctx, req)
}

func (s *Service) ApplySnapshot(ctx context.Context, req syncflow.ApplySnapshotRequest) (syncflow.ApplySnapshotResult, error) {
	data, _, syncService, err := s.loadSyncContext(req.ProfileID)
	if err != nil {
		return syncflow.ApplySnapshotResult{}, err
	}
	if empty(req.MasterPassword) {
		req.MasterPassword = data.MasterPassword
	}
	return syncService.ApplySnapshot(ctx, req)
}

func (s *Service) UploadSync(ctx context.Context) (string, error) {
	data, err := s.store.Load()
	if err != nil {
		return "", err
	}
	if empty(data.ActiveProfileID) {
		return "", syncflow.ErrProfileNotFound
	}
	result, err := s.UploadProfile(ctx, data.ActiveProfileID, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("上传完成，快照 %s，文件 %d 个", result.SnapshotID, result.Uploaded), nil
}

func (s *Service) DownloadSync(ctx context.Context, overwrite bool) (string, error) {
	data, err := s.store.Load()
	if err != nil {
		return "", err
	}
	if empty(data.ActiveProfileID) {
		return "", syncflow.ErrProfileNotFound
	}
	conflicts, err := s.PreviewApplyConflicts(ctx, syncflow.ApplySnapshotRequest{ProfileID: data.ActiveProfileID})
	if err != nil {
		return "", err
	}
	overwriteIDs, overwriteRequired := buildOverwriteList(conflicts, overwrite)
	if overwriteRequired {
		return "", ErrOverwriteRequired
	}
	result, err := s.ApplySnapshot(ctx, syncflow.ApplySnapshotRequest{
		ProfileID:        data.ActiveProfileID,
		OverwriteItemIDs: overwriteIDs,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("下载完成，应用 %d，跳过 %d", result.Applied, result.Skipped), nil
}

func buildOverwriteList(conflicts []syncflow.ApplyConflict, overwrite bool) ([]string, bool) {
	if overwrite {
		ids := make([]string, 0, len(conflicts))
		for _, c := range conflicts {
			ids = append(ids, c.ItemID)
		}
		return ids, false
	}
	if len(conflicts) > 0 {
		return nil, true
	}
	return []string{}, false
}

func (s *Service) loadSyncContext(profileID string) (settings.Data, settings.Profile, SyncService, error) {
	data, err := s.store.Load()
	if err != nil {
		return settings.Data{}, settings.Profile{}, nil, err
	}
	resolvedID := resolveProfileID(data, profileID)
	profile, ok := findProfile(data, resolvedID)
	if !ok {
		return settings.Data{}, settings.Profile{}, nil, syncflow.ErrProfileNotFound
	}
	syncService, err := s.buildSync(data.Token)
	if err != nil {
		return settings.Data{}, settings.Profile{}, nil, err
	}
	return data, *profile, syncService, nil
}
