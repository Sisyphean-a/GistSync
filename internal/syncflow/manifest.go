package syncflow

import (
	"context"
	"encoding/json"
	"fmt"

	"GistSync/internal/settings"
)

func (s *Service) loadManifest(ctx context.Context) (string, manifest, error) {
	gistID, err := s.cloud.EnsureManifestGist(ctx)
	if err != nil {
		return "", manifest{}, err
	}
	content, err := s.cloud.GetFileContent(ctx, FileRequest{GistID: gistID, FileName: manifestFileName})
	if err != nil {
		return "", manifest{}, err
	}
	if empty(content) {
		return gistID, manifest{Version: manifestVersion}, nil
	}
	var data manifest
	if err = json.Unmarshal([]byte(content), &data); err != nil {
		return "", manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	if data.Version == 0 {
		data.Version = manifestVersion
	}
	return gistID, data, nil
}

func (s *Service) saveManifest(ctx context.Context, gistID string, data manifest) error {
	data.Version = manifestVersion
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encode manifest: %w", err)
	}
	return s.cloud.UpsertFile(ctx, UpsertFileRequest{GistID: gistID, FileName: manifestFileName, Content: string(raw)})
}

func (m *manifest) upsertProfile(profile settings.Profile) {
	entry := manifestProfile{ID: profile.ID, Name: profile.Name, RestoreMode: profile.RestoreMode, RestoreRoot: profile.RestoreRoot}
	for _, item := range profile.Items {
		entry.Items = append(entry.Items, manifestProfileItem{
			ID:                 item.ID,
			SourcePathTemplate: item.SourcePathTemplate,
			RelativePath:       normalizeRelative(item),
			Enabled:            item.Enabled,
		})
	}
	for i := range m.Profiles {
		if m.Profiles[i].ID == profile.ID {
			m.Profiles[i] = entry
			return
		}
	}
	m.Profiles = append(m.Profiles, entry)
}

func (m *manifest) findProfileAndSnapshot(profileID string, snapshotID string) (manifestProfile, manifestSnapshot, error) {
	profile, ok := m.findProfile(profileID)
	if !ok {
		return manifestProfile{}, manifestSnapshot{}, ErrProfileNotFound
	}
	snap, ok := m.findSnapshot(profileID, snapshotID)
	if !ok {
		return manifestProfile{}, manifestSnapshot{}, ErrSnapshotNotFound
	}
	return profile, snap, nil
}

func (m *manifest) findProfile(profileID string) (manifestProfile, bool) {
	for _, profile := range m.Profiles {
		if profile.ID == profileID {
			return profile, true
		}
	}
	return manifestProfile{}, false
}

func (m *manifest) findSnapshot(profileID string, snapshotID string) (manifestSnapshot, bool) {
	if !empty(snapshotID) {
		for _, snap := range m.Snapshots {
			if snap.ProfileID == profileID && snap.ID == snapshotID {
				return snap, true
			}
		}
		return manifestSnapshot{}, false
	}
	var latest manifestSnapshot
	found := false
	for _, snap := range m.Snapshots {
		if snap.ProfileID != profileID {
			continue
		}
		if !found || snap.CreatedAt > latest.CreatedAt {
			latest = snap
			found = true
		}
	}
	return latest, found
}
