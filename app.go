package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"GistSync/internal/gistapi"
	"GistSync/internal/settings"
	"GistSync/internal/syncflow"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx      context.Context
	settings *settings.Store
}

func NewApp() *App {
	store, err := settings.NewDefaultStore()
	if err != nil {
		panic(err)
	}
	return &App{settings: store}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) LoadSettingsV2() (settings.Data, error) {
	data, err := a.settings.Load()
	if err != nil {
		return settings.Data{}, err
	}
	if len(data.Profiles) == 0 && strings.TrimSpace(data.Token) != "" && !data.CloudBootstrapDone {
		merged, pullErr := a.pullProfilesIntoSettings(data)
		if pullErr == nil {
			return merged, nil
		}
	}
	return data, nil
}

func (a *App) SaveSettingsV2(data settings.Data) error {
	return a.settings.Save(data)
}

func (a *App) PullProfilesFromCloud() (int, error) {
	data, err := a.settings.Load()
	if err != nil {
		return 0, err
	}
	merged, err := a.pullProfilesIntoSettings(data)
	if err != nil {
		return 0, err
	}
	return len(merged.Profiles), nil
}

func (a *App) CreateProfile(name string) (settings.Profile, error) {
	data, err := a.settings.Load()
	if err != nil {
		return settings.Profile{}, err
	}
	profile := settings.Profile{
		ID:          buildAppID("profile"),
		Name:        strings.TrimSpace(name),
		RestoreMode: "original",
		Enabled:     true,
		Items:       []settings.ProfileItem{},
	}
	data.Profiles = append(data.Profiles, profile)
	if data.ActiveProfileID == "" {
		data.ActiveProfileID = profile.ID
	}
	if err = a.settings.Save(data); err != nil {
		return settings.Profile{}, err
	}
	return profile, nil
}

func (a *App) DeleteProfile(profileID string) error {
	data, err := a.settings.Load()
	if err != nil {
		return err
	}
	filtered := make([]settings.Profile, 0, len(data.Profiles))
	for _, profile := range data.Profiles {
		if profile.ID != profileID {
			filtered = append(filtered, profile)
		}
	}
	data.Profiles = filtered
	if data.ActiveProfileID == profileID {
		data.ActiveProfileID = ""
		if len(data.Profiles) > 0 {
			data.ActiveProfileID = data.Profiles[0].ID
		}
	}
	return a.settings.Save(data)
}

func (a *App) SetActiveProfile(profileID string) error {
	data, err := a.settings.Load()
	if err != nil {
		return err
	}
	if _, ok := findProfile(data, profileID); !ok {
		return syncflow.ErrProfileNotFound
	}
	data.ActiveProfileID = profileID
	return a.settings.Save(data)
}

func (a *App) ChooseFilesForProfile(profileID string) ([]string, error) {
	selected, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择要加入配置的文件（可多选）",
	})
	if err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		return []string{}, nil
	}
	if err = a.AddFilesToProfile(profileID, selected); err != nil {
		return nil, err
	}
	return selected, nil
}

func (a *App) AddFilesToProfile(profileID string, paths []string) error {
	data, err := a.settings.Load()
	if err != nil {
		return err
	}
	profile, ok := findProfile(data, profileID)
	if !ok {
		return syncflow.ErrProfileNotFound
	}
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		item := settings.ProfileItem{
			ID:                 buildAppID("item"),
			SourcePathTemplate: path,
			RelativePath:       buildRelativePath(path),
			Enabled:            true,
		}
		profile.Items = append(profile.Items, item)
	}
	return a.settings.Save(data)
}

func (a *App) RemoveProfileItems(profileID string, itemIDs []string) error {
	data, err := a.settings.Load()
	if err != nil {
		return err
	}
	profile, ok := findProfile(data, profileID)
	if !ok {
		return syncflow.ErrProfileNotFound
	}
	drop := make(map[string]bool, len(itemIDs))
	for _, id := range itemIDs {
		drop[id] = true
	}
	filtered := make([]settings.ProfileItem, 0, len(profile.Items))
	for _, item := range profile.Items {
		if !drop[item.ID] {
			filtered = append(filtered, item)
		}
	}
	profile.Items = filtered
	return a.settings.Save(data)
}

func (a *App) UploadProfile(profileID string, selectedItemIDs []string) (syncflow.UploadProfileResult, error) {
	data, profile, service, err := a.loadContext(profileID)
	if err != nil {
		return syncflow.UploadProfileResult{}, err
	}
	return service.UploadProfile(a.ctx, syncflow.UploadProfileRequest{
		Profile: profile, MasterPassword: data.MasterPassword, SelectedItemIDs: selectedItemIDs,
	})
}

func (a *App) ListSnapshots(profileID string) ([]syncflow.SnapshotMeta, error) {
	data, err := a.settings.Load()
	if err != nil {
		return nil, err
	}
	resolvedID := resolveProfileID(data, profileID)
	if strings.TrimSpace(resolvedID) == "" {
		return []syncflow.SnapshotMeta{}, nil
	}
	client, err := gistapi.NewClient(gistapi.ClientOptions{Token: data.Token})
	if err != nil {
		return nil, err
	}
	service := syncflow.NewService(syncflow.NewGistGateway(client))
	snapshots, err := service.ListSnapshots(a.ctx, resolvedID)
	if errors.Is(err, syncflow.ErrProfileNotFound) {
		return []syncflow.SnapshotMeta{}, nil
	}
	return snapshots, err
}

func (a *App) PreviewApplyConflicts(req syncflow.ApplySnapshotRequest) ([]syncflow.ApplyConflict, error) {
	data, _, service, err := a.loadContext(req.ProfileID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.MasterPassword) == "" {
		req.MasterPassword = data.MasterPassword
	}
	return service.PreviewApplyConflicts(a.ctx, req)
}

func (a *App) ApplySnapshot(req syncflow.ApplySnapshotRequest) (syncflow.ApplySnapshotResult, error) {
	data, _, service, err := a.loadContext(req.ProfileID)
	if err != nil {
		return syncflow.ApplySnapshotResult{}, err
	}
	if strings.TrimSpace(req.MasterPassword) == "" {
		req.MasterPassword = data.MasterPassword
	}
	return service.ApplySnapshot(a.ctx, req)
}

func (a *App) UploadSync() (string, error) {
	data, err := a.settings.Load()
	if err != nil {
		return "", err
	}
	profileID := data.ActiveProfileID
	if strings.TrimSpace(profileID) == "" {
		return "", syncflow.ErrProfileNotFound
	}
	result, err := a.UploadProfile(profileID, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("上传完成，快照 %s，文件 %d 个", result.SnapshotID, result.Uploaded), nil
}

func (a *App) DownloadSync(overwrite bool) (string, error) {
	data, err := a.settings.Load()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(data.ActiveProfileID) == "" {
		return "", syncflow.ErrProfileNotFound
	}
	conflicts, err := a.PreviewApplyConflicts(syncflow.ApplySnapshotRequest{
		ProfileID: data.ActiveProfileID,
	})
	if err != nil {
		return "", err
	}
	overwriteIDs := []string{}
	if overwrite {
		for _, c := range conflicts {
			overwriteIDs = append(overwriteIDs, c.ItemID)
		}
	} else if len(conflicts) > 0 {
		return "", errors.New("OVERWRITE_REQUIRED")
	}
	result, err := a.ApplySnapshot(syncflow.ApplySnapshotRequest{
		ProfileID: data.ActiveProfileID, OverwriteItemIDs: overwriteIDs,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("下载完成，应用 %d，跳过 %d", result.Applied, result.Skipped), nil
}

func buildAppID(prefix string) string {
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), rand.Intn(1000))
}

func buildRelativePath(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.TrimPrefix(path, "{{HOME}}/")
	path = strings.TrimPrefix(path, "/")
	path = strings.ReplaceAll(path, ":", "")
	return path
}

func findProfile(data settings.Data, profileID string) (*settings.Profile, bool) {
	for i := range data.Profiles {
		if data.Profiles[i].ID == profileID {
			return &data.Profiles[i], true
		}
	}
	return nil, false
}

func (a *App) loadContext(profileID string) (settings.Data, settings.Profile, *syncflow.Service, error) {
	data, err := a.settings.Load()
	if err != nil {
		return settings.Data{}, settings.Profile{}, nil, err
	}
	resolvedID := resolveProfileID(data, profileID)
	profile, ok := findProfile(data, resolvedID)
	if !ok {
		return settings.Data{}, settings.Profile{}, nil, syncflow.ErrProfileNotFound
	}
	client, err := gistapi.NewClient(gistapi.ClientOptions{Token: data.Token})
	if err != nil {
		return settings.Data{}, settings.Profile{}, nil, err
	}
	return data, *profile, syncflow.NewService(syncflow.NewGistGateway(client)), nil
}

func resolveProfileID(data settings.Data, requestedID string) string {
	if strings.TrimSpace(requestedID) != "" {
		if _, ok := findProfile(data, requestedID); ok {
			return requestedID
		}
	}
	if strings.TrimSpace(data.ActiveProfileID) != "" {
		if _, ok := findProfile(data, data.ActiveProfileID); ok {
			return data.ActiveProfileID
		}
	}
	if len(data.Profiles) > 0 {
		return data.Profiles[0].ID
	}
	return ""
}

func (a *App) ChooseSyncFile() (string, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择要同步的文件",
	})
	if err != nil {
		return "", err
	}
	return selection, nil
}

func (a *App) pullProfilesIntoSettings(data settings.Data) (settings.Data, error) {
	client, err := gistapi.NewClient(gistapi.ClientOptions{Token: data.Token})
	if err != nil {
		return settings.Data{}, err
	}
	service := syncflow.NewService(syncflow.NewGistGateway(client))
	cloudProfiles, err := service.ListProfilesFromCloud(a.ctx)
	if err != nil {
		return settings.Data{}, err
	}
	byID := make(map[string]int)
	merged := make([]settings.Profile, 0, len(data.Profiles)+len(cloudProfiles))
	for _, profile := range data.Profiles {
		byID[profile.ID] = len(merged)
		merged = append(merged, profile)
	}
	for _, profile := range cloudProfiles {
		if idx, exists := byID[profile.ID]; exists {
			merged[idx] = profile
			continue
		}
		byID[profile.ID] = len(merged)
		merged = append(merged, profile)
	}
	data.Profiles = merged
	data.CloudBootstrapDone = true
	if len(data.Profiles) > 0 && strings.TrimSpace(data.ActiveProfileID) == "" {
		data.ActiveProfileID = data.Profiles[0].ID
	}
	if saveErr := a.settings.Save(data); saveErr != nil {
		return settings.Data{}, saveErr
	}
	return data, nil
}
