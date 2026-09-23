package archive

import (
	"archive/zip"
	"path/filepath"
	"strings"
)

// skipDirs are directories we never treat as project roots and never
// group files under, since they hold vendored/dependency files that
// would otherwise look like their own projects (a package.json inside
// node_modules, for instance).
var skipDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	".git":         true,
	"venv":         true,
	".venv":        true,
	"__pycache__":  true,
	"dist":         true,
	"build":        true,
}

// Tree groups every file in the archive by the directory it lives in.
// e.g. "backend" -> {"requirements.txt": true, "main.py": true}
type Tree map[string]map[string]bool

func Read(zipPath string) (Tree, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	tree := Tree{}
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		dir := filepath.Dir(f.Name)
		if dir == "." {
			dir = ""
		}
		if inSkippedDir(dir) {
			continue
		}
		if tree[dir] == nil {
			tree[dir] = map[string]bool{}
		}
		tree[dir][filepath.Base(f.Name)] = true
	}
	return tree, nil
}

func inSkippedDir(dir string) bool {
	for _, part := range strings.Split(dir, "/") {
		if skipDirs[part] {
			return true
		}
	}
	return false
}
