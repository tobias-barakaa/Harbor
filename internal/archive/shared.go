package archive

import "strings"

// skipDirs are directories we never treat as project roots and never
// group files under, since they hold vendored/dependency files that
// would otherwise look like their own projects.
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

// Tree groups every file found by the directory it lives in, e.g.
// "backend" -> {"requirements.txt": true, "main.py": true}. It's the
// common shape a ZIP read and a local directory walk both produce, so
// detect and inspect don't care which one it came from.
type Tree map[string]map[string]bool

func (t Tree) add(dir, name string) {
	if t[dir] == nil {
		t[dir] = map[string]bool{}
	}
	t[dir][name] = true
}

func inSkippedDir(dir string) bool {
	for _, part := range strings.Split(dir, "/") {
		if skipDirs[part] {
			return true
		}
	}
	return false
}
