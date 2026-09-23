package archive

import (
	"archive/zip"
	"path/filepath"
)

// ReadZip builds a Tree from a .zip archive on disk.
func ReadZip(zipPath string) (Tree, error) {
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
		tree.add(dir, filepath.Base(f.Name))
	}
	return tree, nil
}
