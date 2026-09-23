package archive

import (
	"os"
	"path/filepath"
)

// ReadDir builds a Tree from a local project directory, walking it the
// same way ReadZip walks an archive, so a directory and a zip of that
// same directory produce identical detection results. Skip directories
// are pruned outright (filepath.SkipDir) rather than just filtered,
// since a real node_modules on disk can be huge — no point descending.
func ReadDir(rootPath string) (Tree, error) {
	tree := Tree{}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if path != rootPath && skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}
		dir := filepath.Dir(rel)
		if dir == "." {
			dir = ""
		}
		tree.add(dir, filepath.Base(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tree, nil
}
