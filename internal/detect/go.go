package detect

import "deployer/internal/app"

type goDetector struct{}

func (goDetector) Detect(root string, files map[string]bool) (app.Application, bool) {
	if !files["go.mod"] {
		return app.Application{}, false
	}
	return app.Application{
		ProjectRoot: root,
		Runtime:     app.RuntimeGo,
		Port:        8080,
		Markers:     []string{"go.mod"},
	}, true
}
