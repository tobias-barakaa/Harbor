// python.go
package detect

import "deployer/internal/app"

type pythonDetector struct{}

func (pythonDetector) Detect(root string, files map[string]bool) (app.Application, bool) {
	var markers []string
	if files["requirements.txt"] {
		markers = append(markers, "requirements.txt")
	}
	if files["pyproject.toml"] {
		markers = append(markers, "pyproject.toml")
	}
	if len(markers) == 0 {
		return app.Application{}, false
	}
	return app.Application{
		ProjectRoot: root,
		Runtime:     app.RuntimePython,
		Strategy:    app.StrategyServer,
		Port:        8000,
		Markers:     markers,
	}, true
}