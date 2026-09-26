package app

type Runtime string

const (
	RuntimeNode    Runtime = "Node.js"
	RuntimePython  Runtime = "Python"
	RuntimeGo      Runtime = "Go"
	RuntimeUnknown Runtime = "Unknown"
)

// Framework identifies a specific framework within a runtime, when
// one is recognized. Empty means "plain <Runtime>, no framework
// detected" — not an error, just less specific information.
type Framework string

const (
	FrameworkAstro Framework = "Astro"
	FrameworkNext  Framework = "Next.js"
	FrameworkVite  Framework = "Vite"
)

// Strategy describes how the BUILT application is meant to run —
// which is what actually determines the Dockerfile shape, not the
// language runtime by itself. An Astro app and an Express app are
// both "Node.js", but one produces a directory of static files with
// nothing to execute, and the other is a long-running process.
type Strategy string

const (
	// StrategyServer: the build output IS the running application —
	// a long-lived process listens on Port.
	StrategyServer Strategy = "server"
	// StrategyStatic: the build output is a directory of static
	// files with nothing to execute — any web server can serve it.
	StrategyStatic Strategy = "static"
)

type DeploymentMethod string

const (
	DeploymentDocker   DeploymentMethod = "Docker"
	DeploymentStandard DeploymentMethod = "Standard"
)

type Application struct {
	Name        string
	ProjectRoot string
	Runtime     Runtime
	Framework   Framework // "" if none recognized
	Strategy    Strategy
	BuildCmd    string // "" if there's no separate build step
	OutputDir   string // only meaningful when Strategy == StrategyStatic

	// Port is the external port the deployed application should be
// reachable on. The Docker deployment layer decides which internal
// container port this maps to based on the application's strategy.
Port int

	DeploymentMethod DeploymentMethod
	Markers          []string
}