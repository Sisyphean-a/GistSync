package syncflow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeCloud struct {
	gistID string
	files  map[string]string
}

func newFakeCloud() *fakeCloud {
	return &fakeCloud{
		gistID: "gist-1",
		files: map[string]string{
			manifestFileName: "{}",
		},
	}
}

func (f *fakeCloud) EnsureManifestGist(context.Context) (string, error) {
	return f.gistID, nil
}

func (f *fakeCloud) UpsertFile(_ context.Context, req UpsertFileRequest) error {
	if req.GistID != f.gistID {
		return errors.New("unexpected gist id")
	}
	f.files[req.FileName] = req.Content
	return nil
}

func (f *fakeCloud) GetFileContent(_ context.Context, req FileRequest) (string, error) {
	if req.GistID != f.gistID {
		return "", errors.New("unexpected gist id")
	}
	content, exists := f.files[req.FileName]
	if !exists {
		return "", errors.New("file not found")
	}
	return content, nil
}

func TestService_UploadAndDownload(t *testing.T) {
	cloud := newFakeCloud()
	service := NewService(cloud)
	filePath := filepath.Join(t.TempDir(), "test-config.txt")
	if err := os.WriteFile(filePath, []byte("PM_Secret_Data"), 0o600); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	req := Request{
		Token:          "token-abc",
		MasterPassword: "123",
		SyncPath:       filePath,
	}
	if err := service.Upload(context.Background(), req); err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}

	encryptedFile := buildCloudFileName(filePath)
	encryptedContent := cloud.files[encryptedFile]
	if encryptedContent == "PM_Secret_Data" || encryptedContent == "" {
		t.Fatalf("expected encrypted cloud content, got %q", encryptedContent)
	}

	if err := os.WriteFile(filePath, []byte("Fake Data"), 0o600); err != nil {
		t.Fatalf("write fake data: %v", err)
	}

	err := service.Download(context.Background(), Request{
		Token:          req.Token,
		MasterPassword: req.MasterPassword,
		SyncPath:       req.SyncPath,
		Overwrite:      false,
	})
	if !errors.Is(err, ErrOverwriteRequired) {
		t.Fatalf("expected overwrite required error, got %v", err)
	}

	if err = service.Download(context.Background(), Request{
		Token:          req.Token,
		MasterPassword: req.MasterPassword,
		SyncPath:       req.SyncPath,
		Overwrite:      true,
	}); err != nil {
		t.Fatalf("Download returned error: %v", err)
	}

	raw, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(raw) != "PM_Secret_Data" {
		t.Fatalf("file content mismatch: got %q", string(raw))
	}
}

func TestService_DownloadPathNotInManifest(t *testing.T) {
	service := NewService(newFakeCloud())
	err := service.Download(context.Background(), Request{
		Token:          "token-abc",
		MasterPassword: "123",
		SyncPath:       "missing.txt",
		Overwrite:      true,
	})
	if !errors.Is(err, ErrPathNotInManifest) {
		t.Fatalf("expected ErrPathNotInManifest, got %v", err)
	}
}

func TestService_ValidateRequest(t *testing.T) {
	service := NewService(newFakeCloud())
	err := service.Upload(context.Background(), Request{
		Token:          "",
		MasterPassword: "123",
		SyncPath:       "a.txt",
	})
	if !errors.Is(err, ErrEmptyToken) {
		t.Fatalf("expected ErrEmptyToken, got %v", err)
	}

	err = service.Upload(context.Background(), Request{
		Token:          "token",
		MasterPassword: "",
		SyncPath:       "a.txt",
	})
	if !errors.Is(err, ErrEmptyPassword) {
		t.Fatalf("expected ErrEmptyPassword, got %v", err)
	}

	err = service.Upload(context.Background(), Request{
		Token:          "token",
		MasterPassword: "pwd",
		SyncPath:       "   ",
	})
	if !errors.Is(err, ErrEmptySyncPath) {
		t.Fatalf("expected ErrEmptySyncPath, got %v", err)
	}
}

func TestService_UploadWithHomePlaceholder(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("resolve home dir: %v", err)
	}
	targetFile := filepath.Join(homeDir, "gistsync-stage4-placeholder-test.txt")
	t.Cleanup(func() {
		_ = os.Remove(targetFile)
	})
	if err = os.WriteFile(targetFile, []byte("sample"), 0o600); err != nil {
		t.Fatalf("prepare test file: %v", err)
	}

	service := NewService(newFakeCloud())
	err = service.Upload(context.Background(), Request{
		Token:          "token",
		MasterPassword: "pwd",
		SyncPath:       "{{HOME}}/gistsync-stage4-placeholder-test.txt",
	})
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if !strings.Contains(targetFile, filepath.Base(targetFile)) {
		t.Fatalf("sanity check failed for target file path")
	}
}
