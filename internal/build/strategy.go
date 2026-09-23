package build

import "deployer/internal/app"

// Result describes the outcome of building an Application in a
// prepared workspace directory.
type Result struct {
	Success   bool
	StartCmd  string   // executable to launch the app after build
	StartArgs []string // arguments for StartCmd
	Log       string   // combined build output, for the caller to display or store
}

// Strategy knows how to turn one runtime's source tree into something
// runnable, and how to start it once built.
type Strategy interface {
	Name() string
	CanHandle(a app.Application) bool
	Build(workDir string, a app.Application) (Result, error)
}

// Registry holds every build strategy, tried in order.
var Registry = []Strategy{
	nodeStrategy{},
	pythonStrategy{},
}

// For finds the first strategy that can handle the given Application.
func For(a app.Application) (Strategy, bool) {
	for _, s := range Registry {
		if s.CanHandle(a) {
			return s, true
		}
	}
	return nil, false
}
