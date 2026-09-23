package detect

import "deployer/internal/app"

type nodeDetector struct{}

func (nodeDetector) Detect(root string, files map[string]bool) (app.Application, bool) {
	if !files["package.json"] {
		return app.Application{}, false
	}
	return app.Application{
		ProjectRoot: root,
		Runtime:     app.RuntimeNode,
		Port:        3000,
		Markers:     []string{"package.json"},
	}, true
}
