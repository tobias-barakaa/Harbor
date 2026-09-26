package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"deployer/internal/inspect"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

type AppInfo struct {
	Name      string `json:"name"`
	Runtime   string `json:"runtime"`
	Framework string `json:"framework"`
	Strategy  string `json:"strategy"`
	Port      int    `json:"port"`
}

func (a *App) InspectPath(path string) ([]AppInfo, error) {
	apps, err := inspect.Path(path)
	if err != nil {
		return nil, err
	}
	result := make([]AppInfo, 0, len(apps))
	for _, ap := range apps {
		result = append(result, AppInfo{
			Name:      ap.Name,
			Runtime:   string(ap.Runtime),
			Framework: string(ap.Framework),
			Strategy:  string(ap.Strategy),
			Port:      ap.Port,
		})
	}
	return result, nil
}

func (a *App) PickDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select application directory",
	})
}

func (a *App) PickZipFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select application zip",
		Filters: []runtime.FileFilter{
			{DisplayName: "Zip files (*.zip)", Pattern: "*.zip"},
		},
	})
}