package detect

import "deployer/internal/app"

type nodeDetector struct{}

func (nodeDetector) Detect(root string, files map[string]bool) (app.Application, bool) {
	if !files["package.json"] {
		return app.Application{}, false
	}
	markers := []string{"package.json"}

	if fw, ok := detectNodeFramework(files); ok {
		markers = append(markers, fw.Marker)
		return app.Application{
			ProjectRoot: root,
			Runtime:     app.RuntimeNode,
			Framework:   fw.Framework,
			Strategy:    fw.Strategy,
			BuildCmd:    fw.BuildCmd,
			OutputDir:   fw.OutputDir,
			Port:        fw.Port,
			Markers:     markers,
		}, true
	}

	// No recognized framework — plain Node.js, same defaults as
	// before framework detection existed.
	return app.Application{
		ProjectRoot: root,
		Runtime:     app.RuntimeNode,
		Strategy:    app.StrategyServer,
		Port:        3000,
		Markers:     markers,
	}, true
}