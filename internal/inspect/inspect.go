package inspect

import (
	"os"
	"path/filepath"

	"deployer/internal/app"
	"deployer/internal/archive"
	"deployer/internal/detect"
)

// Path inspects either a local project directory or a .zip archive,
// picking the reader based on what target actually is, then runs
// every registered detector against every directory found.
func Path(target string) ([]app.Application, error) {
	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}

	var tree archive.Tree
	if info.IsDir() {
		tree, err = archive.ReadDir(target)
	} else {
		tree, err = archive.ReadZip(target)
	}
	if err != nil {
		return nil, err
	}

	return fromTree(tree, target), nil
}

func fromTree(tree archive.Tree, target string) []app.Application {
	var apps []app.Application
	for dir, files := range tree {
		for _, d := range detect.Registry {
			result, ok := d.Detect(dir, files)
			if !ok {
				continue
			}
			result.Name = nameFor(dir, target)
			result.DeploymentMethod = app.DeploymentStandard
			if files["Dockerfile"] {
				result.DeploymentMethod = app.DeploymentDocker
			}
			apps = append(apps, result)
			break
		}
	}
	return apps
}

func nameFor(dir, target string) string {
	if dir != "" {
		return filepath.Base(dir)
	}
	base := filepath.Base(target)
	if ext := filepath.Ext(base); ext == ".zip" {
		return base[:len(base)-len(ext)]
	}
	return base
}
