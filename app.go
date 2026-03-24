package main

import (
	"context"
	"errors"
	"fmt"

	"GistSync/internal/gistapi"
	"GistSync/internal/settings"
	"GistSync/internal/syncflow"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx      context.Context
	settings *settings.Store
}

// NewApp creates a new App application struct
func NewApp() *App {
	store, err := settings.NewDefaultStore()
	if err != nil {
		panic(err)
	}
	return &App{settings: store}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) LoadSettings() (settings.Data, error) {
	return a.settings.Load()
}

func (a *App) SaveSettings(data settings.Data) error {
	return a.settings.Save(data)
}

func (a *App) UploadSync() (string, error) {
	data, err := a.settings.Load()
	if err != nil {
		return "", err
	}
	service, err := newSyncService(data.Token)
	if err != nil {
		return "", err
	}
	err = service.Upload(a.ctx, syncflow.Request{
		Token:          data.Token,
		MasterPassword: data.MasterPassword,
		SyncPath:       data.SyncPath,
	})
	if err != nil {
		return "", err
	}
	return "上传同步完成", nil
}

func (a *App) DownloadSync(overwrite bool) (string, error) {
	data, err := a.settings.Load()
	if err != nil {
		return "", err
	}
	service, err := newSyncService(data.Token)
	if err != nil {
		return "", err
	}
	err = service.Download(a.ctx, syncflow.Request{
		Token:          data.Token,
		MasterPassword: data.MasterPassword,
		SyncPath:       data.SyncPath,
		Overwrite:      overwrite,
	})
	if errors.Is(err, syncflow.ErrOverwriteRequired) {
		return "", fmt.Errorf("OVERWRITE_REQUIRED: %w", err)
	}
	if err != nil {
		return "", err
	}
	return "下载同步完成", nil
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

func newSyncService(token string) (*syncflow.Service, error) {
	client, err := gistapi.NewClient(gistapi.ClientOptions{Token: token})
	if err != nil {
		return nil, err
	}
	return syncflow.NewService(syncflow.NewGistGateway(client)), nil
}
