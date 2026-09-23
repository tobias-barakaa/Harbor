package inspect

import (
	"path/filepath"

	"deployer/internal/app"
	"deployer/internal/archive"
	"deployer/internal/detect"
)

// Zip runs every registered detector against every directory found in
// the archive, returning one Application per directory that matched.
// This is what lets a zip bundling a frontend and backend come back
// as multiple Applications instead of a single guess.
func Zip(zipPath string) ([]app.Application, error) {
	tree, err := archive.Read(zipPath)
	if err != nil {
		return nil, err
	}

	var apps []app.Application
	for dir, files := range tree {
		for _, d := range detect.Registry {
			result, ok := d.Detect(dir, files)
			if !ok {
				continue
			}
			result.Name = nameFor(dir, zipPath)
			result.DeploymentMethod = app.DeploymentStandard
			if files["Dockerfile"] {
				result.DeploymentMethod = app.DeploymentDocker
			}
			apps = append(apps, result)
			break
		}
	}
	return apps, nil
}

func nameFor(dir, zipPath string) string {
	if dir != "" {
		return filepath.Base(dir)
	}
	base := filepath.Base(zipPath)
	return base[:len(base)-len(filepath.Ext(base))]
}
