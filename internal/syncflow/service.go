package syncflow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"GistSync/internal/pathmap"
	"GistSync/internal/security"
	"GistSync/internal/settings"
)

const (
	manifestFileName = "sync_manifest.json"
	manifestVersion  = 2
	restoreOriginal  = "original"
	restoreRooted    = "rooted"
)

var (
	ErrProfileNotFound  = errors.New("profile not found")
	ErrSnapshotNotFound = errors.New("snapshot not found")
	ErrEmptyPassword    = errors.New("master password is required")
	ErrEmptyRestoreRoot = errors.New("restore root is required when mode is rooted")
)

type CloudGateway interface {
	EnsureManifestGist(ctx context.Context) (string, error)
	UpsertFile(ctx context.Context, req UpsertFileRequest) error
	GetFileContent(ctx context.Context, req FileRequest) (string, error)
}

type UpsertFileRequest struct {
	GistID   string
	FileName string
	Content  string
}

type FileRequest struct {
	GistID   string
	FileName string
}

type UploadProfileRequest struct {
	Profile        settings.Profile
	MasterPassword string
}

type UploadProfileResult struct {
	SnapshotID string `json:"snapshotId"`
	Uploaded   int    `json:"uploaded"`
}

type SnapshotMeta struct {
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`
}

type ApplyConflict struct {
	ItemID     string `json:"itemId"`
	TargetPath string `json:"targetPath"`
}

type ApplySnapshotRequest struct {
	ProfileID        string   `json:"profileId"`
	SnapshotID       string   `json:"snapshotId"`
	MasterPassword   string   `json:"masterPassword"`
	RestoreMode      string   `json:"restoreMode"`
	RestoreRoot      string   `json:"restoreRoot"`
	OverwriteItemIDs []string `json:"overwriteItemIds"`
}

type ApplyItemResult struct {
	ItemID     string `json:"itemId"`
	TargetPath string `json:"targetPath"`
	Status     string `json:"status"`
	Reason     string `json:"reason"`
}

type ApplySnapshotResult struct {
	Applied int               `json:"applied"`
	Skipped int               `json:"skipped"`
	Items   []ApplyItemResult `json:"items"`
}

type Service struct {
	cloud CloudGateway
}

type manifest struct {
	Version   int                `json:"version"`
	Profiles  []manifestProfile  `json:"profiles"`
	Snapshots []manifestSnapshot `json:"snapshots"`
}

type manifestProfile struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	RestoreMode string                `json:"restoreMode"`
	RestoreRoot string                `json:"restoreRoot"`
	Items       []manifestProfileItem `json:"items"`
}

type manifestProfileItem struct {
	ID                 string `json:"id"`
	SourcePathTemplate string `json:"sourcePathTemplate"`
	RelativePath       string `json:"relativePath"`
	Enabled            bool   `json:"enabled"`
}

type manifestSnapshot struct {
	ID        string                 `json:"id"`
	ProfileID string                 `json:"profileId"`
	CreatedAt string                 `json:"createdAt"`
	Items     []manifestSnapshotItem `json:"items"`
}

type manifestSnapshotItem struct {
	ItemID             string `json:"itemId"`
	SourcePathTemplate string `json:"sourcePathTemplate"`
	RelativePath       string `json:"relativePath"`
	BlobFile           string `json:"blobFile"`
}

func NewService(cloud CloudGateway) *Service {
	return &Service{cloud: cloud}
}

func (s *Service) ListProfilesFromCloud(ctx context.Context) ([]settings.Profile, error) {
	_, data, err := s.loadManifest(ctx)
	if err != nil {
		return nil, err
	}
	profiles := make([]settings.Profile, 0, len(data.Profiles))
	for _, profile := range data.Profiles {
		items := make([]settings.ProfileItem, 0, len(profile.Items))
		for _, item := range profile.Items {
			items = append(items, settings.ProfileItem{
				ID:                 item.ID,
				SourcePathTemplate: item.SourcePathTemplate,
				RelativePath:       item.RelativePath,
				Enabled:            item.Enabled,
			})
		}
		profiles = append(profiles, settings.Profile{
			ID:          profile.ID,
			Name:        profile.Name,
			RestoreMode: profile.RestoreMode,
			RestoreRoot: profile.RestoreRoot,
			Enabled:     true,
			Items:       items,
		})
	}
	return profiles, nil
}

func (s *Service) UploadProfile(ctx context.Context, req UploadProfileRequest) (UploadProfileResult, error) {
	if strings.TrimSpace(req.MasterPassword) == "" {
		return UploadProfileResult{}, ErrEmptyPassword
	}
	gistID, data, err := s.loadManifest(ctx)
	if err != nil {
		return UploadProfileResult{}, err
	}

	snapshot := manifestSnapshot{
		ID:        buildID("snapshot", req.Profile.ID),
		ProfileID: req.Profile.ID,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	for _, item := range req.Profile.Items {
		if !item.Enabled {
			continue
		}
		absolutePath, resolveErr := pathmap.ExpandHomePath(item.SourcePathTemplate)
		if resolveErr != nil {
			return UploadProfileResult{}, resolveErr
		}
		raw, readErr := os.ReadFile(absolutePath)
		if readErr != nil {
			return UploadProfileResult{}, fmt.Errorf("read local file: %w", readErr)
		}
		encrypted, encErr := security.EncryptString(string(raw), req.MasterPassword)
		if encErr != nil {
			return UploadProfileResult{}, encErr
		}
		blob := buildBlobFileName(req.Profile.ID, item.ID, time.Now().UnixNano())
		if err = s.cloud.UpsertFile(ctx, UpsertFileRequest{
			GistID: gistID, FileName: blob, Content: encrypted,
		}); err != nil {
			return UploadProfileResult{}, err
		}
		snapshot.Items = append(snapshot.Items, manifestSnapshotItem{
			ItemID:             item.ID,
			SourcePathTemplate: item.SourcePathTemplate,
			RelativePath:       normalizeRelative(item),
			BlobFile:           blob,
		})
	}

	data.upsertProfile(req.Profile)
	data.Snapshots = append(data.Snapshots, snapshot)
	if err = s.saveManifest(ctx, gistID, data); err != nil {
		return UploadProfileResult{}, err
	}
	return UploadProfileResult{SnapshotID: snapshot.ID, Uploaded: len(snapshot.Items)}, nil
}

func (s *Service) ListSnapshots(ctx context.Context, profileID string) ([]SnapshotMeta, error) {
	_, data, err := s.loadManifest(ctx)
	if err != nil {
		return nil, err
	}
	var out []SnapshotMeta
	for _, snap := range data.Snapshots {
		if snap.ProfileID != profileID {
			continue
		}
		out = append(out, SnapshotMeta{ID: snap.ID, CreatedAt: snap.CreatedAt})
	}
	sort.Slice(out, func(i int, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	return out, nil
}

func (s *Service) PreviewApplyConflicts(ctx context.Context, req ApplySnapshotRequest) ([]ApplyConflict, error) {
	_, data, err := s.loadManifest(ctx)
	if err != nil {
		return nil, err
	}
	profile, snap, err := data.findProfileAndSnapshot(req.ProfileID, req.SnapshotID)
	if err != nil {
		return nil, err
	}
	restoreMode := chooseRestoreMode(req.RestoreMode, profile.RestoreMode)
	if restoreMode == restoreRooted && strings.TrimSpace(req.RestoreRoot) == "" {
		return nil, ErrEmptyRestoreRoot
	}

	var conflicts []ApplyConflict
	for _, item := range snap.Items {
		targetPath, targetErr := resolveTargetPath(item, restoreMode, req.RestoreRoot)
		if targetErr != nil {
			return nil, targetErr
		}
		if _, statErr := os.Stat(targetPath); statErr == nil {
			conflicts = append(conflicts, ApplyConflict{
				ItemID: item.ItemID, TargetPath: targetPath,
			})
		}
	}
	return conflicts, nil
}

func (s *Service) ApplySnapshot(ctx context.Context, req ApplySnapshotRequest) (ApplySnapshotResult, error) {
	if strings.TrimSpace(req.MasterPassword) == "" {
		return ApplySnapshotResult{}, ErrEmptyPassword
	}
	gistID, data, err := s.loadManifest(ctx)
	if err != nil {
		return ApplySnapshotResult{}, err
	}
	profile, snap, err := data.findProfileAndSnapshot(req.ProfileID, req.SnapshotID)
	if err != nil {
		return ApplySnapshotResult{}, err
	}
	restoreMode := chooseRestoreMode(req.RestoreMode, profile.RestoreMode)
	if restoreMode == restoreRooted && strings.TrimSpace(req.RestoreRoot) == "" {
		return ApplySnapshotResult{}, ErrEmptyRestoreRoot
	}

	overwriteSet := make(map[string]bool, len(req.OverwriteItemIDs))
	for _, id := range req.OverwriteItemIDs {
		overwriteSet[id] = true
	}

	result := ApplySnapshotResult{}
	for _, item := range snap.Items {
		itemResult := ApplyItemResult{ItemID: item.ItemID}
		targetPath, targetErr := resolveTargetPath(item, restoreMode, req.RestoreRoot)
		if targetErr != nil {
			itemResult.Status = "error"
			itemResult.Reason = targetErr.Error()
			result.Items = append(result.Items, itemResult)
			result.Skipped++
			continue
		}
		itemResult.TargetPath = targetPath
		if shouldSkip(targetPath, overwriteSet[item.ItemID]) {
			itemResult.Status = "skipped"
			itemResult.Reason = "target exists and overwrite not granted"
			result.Items = append(result.Items, itemResult)
			result.Skipped++
			continue
		}

		encrypted, readErr := s.cloud.GetFileContent(ctx, FileRequest{GistID: gistID, FileName: item.BlobFile})
		if readErr != nil {
			itemResult.Status = "error"
			itemResult.Reason = readErr.Error()
			result.Items = append(result.Items, itemResult)
			result.Skipped++
			continue
		}
		decrypted, decErr := security.DecryptString(encrypted, req.MasterPassword)
		if decErr != nil {
			itemResult.Status = "error"
			itemResult.Reason = decErr.Error()
			result.Items = append(result.Items, itemResult)
			result.Skipped++
			continue
		}
		if writeErr := writeFile(targetPath, decrypted); writeErr != nil {
			itemResult.Status = "error"
			itemResult.Reason = writeErr.Error()
			result.Items = append(result.Items, itemResult)
			result.Skipped++
			continue
		}
		itemResult.Status = "applied"
		result.Items = append(result.Items, itemResult)
		result.Applied++
	}
	return result, nil
}

func shouldSkip(path string, overwrite bool) bool {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return err == nil && !overwrite
}

func writeFile(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write target file: %w", err)
	}
	return nil
}

func resolveTargetPath(item manifestSnapshotItem, mode string, restoreRoot string) (string, error) {
	if mode == restoreRooted {
		if strings.TrimSpace(restoreRoot) == "" {
			return "", ErrEmptyRestoreRoot
		}
		return filepath.Join(restoreRoot, filepath.FromSlash(item.RelativePath)), nil
	}
	return pathmap.ExpandHomePath(item.SourcePathTemplate)
}

func normalizeRelative(item settings.ProfileItem) string {
	if strings.TrimSpace(item.RelativePath) != "" {
		return item.RelativePath
	}
	path := strings.ReplaceAll(item.SourcePathTemplate, "\\", "/")
	path = strings.TrimPrefix(path, "{{HOME}}/")
	path = strings.TrimPrefix(path, "/")
	path = strings.ReplaceAll(path, ":", "")
	return path
}

func chooseRestoreMode(requested string, profileDefault string) string {
	if requested == restoreRooted || requested == restoreOriginal {
		return requested
	}
	if profileDefault == restoreRooted {
		return restoreRooted
	}
	return restoreOriginal
}

func (s *Service) loadManifest(ctx context.Context) (string, manifest, error) {
	gistID, err := s.cloud.EnsureManifestGist(ctx)
	if err != nil {
		return "", manifest{}, err
	}
	content, err := s.cloud.GetFileContent(ctx, FileRequest{GistID: gistID, FileName: manifestFileName})
	if err != nil {
		return "", manifest{}, err
	}
	if strings.TrimSpace(content) == "" {
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
	return s.cloud.UpsertFile(ctx, UpsertFileRequest{
		GistID: gistID, FileName: manifestFileName, Content: string(raw),
	})
}

func (m *manifest) upsertProfile(profile settings.Profile) {
	entry := manifestProfile{
		ID:          profile.ID,
		Name:        profile.Name,
		RestoreMode: profile.RestoreMode,
		RestoreRoot: profile.RestoreRoot,
	}
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
	if strings.TrimSpace(snapshotID) != "" {
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

func buildBlobFileName(profileID string, itemID string, now int64) string {
	key := fmt.Sprintf("%s|%s|%d", profileID, itemID, now)
	sum := sha256.Sum256([]byte(key))
	return "blob_" + hex.EncodeToString(sum[:]) + ".enc"
}

func buildID(prefix string, profileID string) string {
	return fmt.Sprintf("%s-%s-%d", prefix, profileID, time.Now().UnixNano())
}
