package workspace

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"deployer/internal/app"
)

// Prepare returns a real filesystem directory containing the
// application's full source tree, ready for build commands to run
// against, plus a cleanup func the caller should defer.
//
// If target is already a directory, this is just target+ProjectRoot —
// no copying needed. If target is a zip, the relevant subtree (every
// entry under ProjectRoot, recursively — Tree only ever recorded
// top-level filenames per directory, not nested files) is extracted
// into a temp directory.
func Prepare(target string, a app.Application) (dir string, cleanup func(), err error) {
	info, err := os.Stat(target)
	if err != nil {
		return "", nil, err
	}
	if info.IsDir() {
		return filepath.Join(target, a.ProjectRoot), func() {}, nil
	}
	return extractZip(target, a.ProjectRoot)
}

func extractZip(zipPath, projectRoot string) (string, func(), error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", nil, err
	}
	defer r.Close()

	tmpDir, err := os.MkdirTemp("", "deployer-workspace-*")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { os.RemoveAll(tmpDir) }

	prefix := projectRoot
	if prefix != "" {
		prefix += "/"
	}

	for _, f := range r.File {
		if prefix != "" && !strings.HasPrefix(f.Name, prefix) {
			continue
		}
		rel := strings.TrimPrefix(f.Name, prefix)
		if rel == "" {
			continue
		}
		destPath := filepath.Join(tmpDir, rel)

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, 0o755)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			cleanup()
			return "", nil, err
		}
		if err := extractFile(f, destPath); err != nil {
			cleanup()
			return "", nil, err
		}
	}
	return tmpDir, cleanup, nil
}

func extractFile(f *zip.File, destPath string) error {
	src, err := f.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
