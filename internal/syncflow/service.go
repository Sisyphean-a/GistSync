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
	"strings"

	"GistSync/internal/pathmap"
	"GistSync/internal/security"
)

const manifestFileName = "sync_manifest.json"

var (
	ErrPathNotInManifest = errors.New("path not found in manifest")
	ErrOverwriteRequired = errors.New("target file already exists, overwrite confirmation required")
	ErrEmptyToken        = errors.New("github token is required")
	ErrEmptyPassword     = errors.New("master password is required")
	ErrEmptySyncPath     = errors.New("sync path is required")
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

type Request struct {
	Token          string
	MasterPassword string
	SyncPath       string
	Overwrite      bool
}

type Service struct {
	cloud CloudGateway
}

type manifest struct {
	Files []manifestItem `json:"files"`
}

type manifestItem struct {
	Path     string `json:"path"`
	GistFile string `json:"gistFile"`
}

func NewService(cloud CloudGateway) *Service {
	return &Service{cloud: cloud}
}

func (s *Service) Upload(ctx context.Context, req Request) error {
	if err := validateRequest(req); err != nil {
		return err
	}

	absolutePath, err := pathmap.ExpandHomePath(req.SyncPath)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(absolutePath)
	if err != nil {
		return fmt.Errorf("read local file: %w", err)
	}
	encrypted, err := security.EncryptString(string(content), req.MasterPassword)
	if err != nil {
		return err
	}
	return s.uploadToCloud(ctx, req.SyncPath, encrypted)
}

func (s *Service) Download(ctx context.Context, req Request) error {
	if err := validateRequest(req); err != nil {
		return err
	}

	absolutePath, err := pathmap.ExpandHomePath(req.SyncPath)
	if err != nil {
		return err
	}
	if err = checkOverwrite(absolutePath, req.Overwrite); err != nil {
		return err
	}
	encrypted, err := s.readEncryptedContent(ctx, req.SyncPath)
	if err != nil {
		return err
	}
	decrypted, err := security.DecryptString(encrypted, req.MasterPassword)
	if err != nil {
		return err
	}
	return writeLocalFile(absolutePath, decrypted)
}

func validateRequest(req Request) error {
	if strings.TrimSpace(req.Token) == "" {
		return ErrEmptyToken
	}
	if strings.TrimSpace(req.MasterPassword) == "" {
		return ErrEmptyPassword
	}
	if strings.TrimSpace(req.SyncPath) == "" {
		return ErrEmptySyncPath
	}
	return nil
}

func checkOverwrite(path string, overwrite bool) error {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect local file: %w", err)
	}
	if overwrite {
		return nil
	}
	return ErrOverwriteRequired
}

func writeLocalFile(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write target file: %w", err)
	}
	return nil
}

func (s *Service) readEncryptedContent(ctx context.Context, syncPath string) (string, error) {
	gistID, manifestData, err := s.loadManifest(ctx)
	if err != nil {
		return "", err
	}
	entry, found := manifestData.findByPath(syncPath)
	if !found {
		return "", ErrPathNotInManifest
	}
	return s.cloud.GetFileContent(ctx, FileRequest{GistID: gistID, FileName: entry.GistFile})
}

func (s *Service) uploadToCloud(ctx context.Context, syncPath string, encrypted string) error {
	gistID, manifestData, err := s.loadManifest(ctx)
	if err != nil {
		return err
	}
	entry := manifestData.upsert(syncPath)
	if err = s.cloud.UpsertFile(ctx, UpsertFileRequest{
		GistID: gistID, FileName: entry.GistFile, Content: encrypted,
	}); err != nil {
		return err
	}

	rawManifest, err := json.Marshal(manifestData)
	if err != nil {
		return fmt.Errorf("encode manifest: %w", err)
	}
	return s.cloud.UpsertFile(ctx, UpsertFileRequest{
		GistID: gistID, FileName: manifestFileName, Content: string(rawManifest),
	})
}

func (s *Service) loadManifest(ctx context.Context) (string, manifest, error) {
	gistID, err := s.cloud.EnsureManifestGist(ctx)
	if err != nil {
		return "", manifest{}, err
	}
	content, err := s.cloud.GetFileContent(ctx, FileRequest{
		GistID: gistID, FileName: manifestFileName,
	})
	if err != nil {
		return "", manifest{}, err
	}
	manifestData, err := parseManifest(content)
	if err != nil {
		return "", manifest{}, err
	}
	return gistID, manifestData, nil
}

func parseManifest(content string) (manifest, error) {
	var data manifest
	if strings.TrimSpace(content) == "" {
		return data, nil
	}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	return data, nil
}

func (m *manifest) findByPath(syncPath string) (manifestItem, bool) {
	for _, item := range m.Files {
		if item.Path == syncPath {
			return item, true
		}
	}
	return manifestItem{}, false
}

func (m *manifest) upsert(syncPath string) manifestItem {
	for i, item := range m.Files {
		if item.Path == syncPath {
			return m.Files[i]
		}
	}
	newItem := manifestItem{
		Path:     syncPath,
		GistFile: buildCloudFileName(syncPath),
	}
	m.Files = append(m.Files, newItem)
	return newItem
}

func buildCloudFileName(syncPath string) string {
	sum := sha256.Sum256([]byte(syncPath))
	return "file_" + hex.EncodeToString(sum[:]) + ".enc"
}
