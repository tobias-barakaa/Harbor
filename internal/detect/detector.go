package detect

import "deployer/internal/app"

// Detector inspects the files found directly in one directory of the
// archive and reports whether it recognizes a project there.
type Detector interface {
	Detect(projectRoot string, files map[string]bool) (app.Application, bool)
}

// Registry holds every detector, tried in order per directory. The
// first match wins for that directory.
var Registry = []Detector{
	nodeDetector{},
	pythonDetector{},
	goDetector{},
}
