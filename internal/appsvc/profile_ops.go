package appsvc

import (
	"context"
	"strings"

	"GistSync/internal/profileutil"
	"GistSync/internal/settings"
	"GistSync/internal/syncflow"
)

func (s *Service) LoadSettings(ctx context.Context) (settings.Data, error) {
	data, err := s.store.Load()
	if err != nil {
		return settings.Data{}, err
	}
	if !shouldBootstrap(data) {
		return data, nil
	}
	merged, pullErr := s.pullProfilesIntoSettings(ctx, data)
	if pullErr == nil {
		return merged, nil
	}
	return data, nil
}

func (s *Service) SaveSettings(_ context.Context, data settings.Data) error {
	return s.store.Save(data)
}

func (s *Service) PullProfilesFromCloud(ctx context.Context) (int, error) {
	data, err := s.store.Load()
	if err != nil {
		return 0, err
	}
	merged, err := s.pullProfilesIntoSettings(ctx, data)
	if err != nil {
		return 0, err
	}
	return len(merged.Profiles), nil
}

func (s *Service) CreateProfile(_ context.Context, name string) (settings.Profile, error) {
	data, err := s.store.Load()
	if err != nil {
		return settings.Profile{}, err
	}
	profile := settings.Profile{
		ID:          s.generateID("profile"),
		Name:        strings.TrimSpace(name),
		RestoreMode: "original",
		Enabled:     true,
		Items:       []settings.ProfileItem{},
	}
	data.Profiles = append(data.Profiles, profile)
	if data.ActiveProfileID == "" {
		data.ActiveProfileID = profile.ID
	}
	if err = s.store.Save(data); err != nil {
		return settings.Profile{}, err
	}
	return profile, nil
}

func (s *Service) DeleteProfile(_ context.Context, profileID string) error {
	data, err := s.store.Load()
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
	return s.store.Save(data)
}

func (s *Service) SetActiveProfile(_ context.Context, profileID string) error {
	data, err := s.store.Load()
	if err != nil {
		return err
	}
	if _, ok := findProfile(data, profileID); !ok {
		return syncflow.ErrProfileNotFound
	}
	data.ActiveProfileID = profileID
	return s.store.Save(data)
}

func (s *Service) AddFilesToProfile(_ context.Context, profileID string, paths []string) error {
	data, err := s.store.Load()
	if err != nil {
		return err
	}
	profile, ok := findProfile(data, profileID)
	if !ok {
		return syncflow.ErrProfileNotFound
	}
	for _, path := range paths {
		if empty(path) {
			continue
		}
		profile.Items = append(profile.Items, settings.ProfileItem{
			ID:                 s.generateID("item"),
			SourcePathTemplate: path,
			RelativePath:       profileutil.NormalizeRelativePath(path),
			Enabled:            true,
		})
	}
	return s.store.Save(data)
}

func (s *Service) RemoveProfileItems(_ context.Context, profileID string, itemIDs []string) error {
	data, err := s.store.Load()
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
	return s.store.Save(data)
}

func (s *Service) pullProfilesIntoSettings(ctx context.Context, data settings.Data) (settings.Data, error) {
	syncService, err := s.buildSync(data.Token)
	if err != nil {
		return settings.Data{}, err
	}
	cloudProfiles, err := syncService.ListProfilesFromCloud(ctx)
	if err != nil {
		return settings.Data{}, err
	}
	data.Profiles = mergeProfiles(data.Profiles, cloudProfiles)
	data.CloudBootstrapDone = true
	if len(data.Profiles) > 0 && empty(data.ActiveProfileID) {
		data.ActiveProfileID = data.Profiles[0].ID
	}
	if saveErr := s.store.Save(data); saveErr != nil {
		return settings.Data{}, saveErr
	}
	return data, nil
}

func mergeProfiles(local []settings.Profile, cloud []settings.Profile) []settings.Profile {
	byID := make(map[string]int)
	merged := make([]settings.Profile, 0, len(local)+len(cloud))
	for _, profile := range local {
		byID[profile.ID] = len(merged)
		merged = append(merged, profile)
	}
	for _, profile := range cloud {
		if idx, exists := byID[profile.ID]; exists {
			merged[idx] = profile
			continue
		}
		byID[profile.ID] = len(merged)
		merged = append(merged, profile)
	}
	return merged
}
