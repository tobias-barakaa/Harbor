package app

type Runtime string

const (
	RuntimeNode    Runtime = "Node.js"
	RuntimePython  Runtime = "Python"
	RuntimeGo      Runtime = "Go"
	RuntimeUnknown Runtime = "Unknown"
)

type DeploymentMethod string

const (
	DeploymentDocker   DeploymentMethod = "Docker"
	DeploymentStandard DeploymentMethod = "Standard"
)

// Application describes one detected project inside an inspected archive.
// A single zip can produce more than one of these (e.g. a frontend and
// a backend living in separate subdirectories).
type Application struct {
	Name             string
	ProjectRoot      string // directory inside the archive; "" means the zip root
	Runtime          Runtime
	Port             int
	DeploymentMethod DeploymentMethod
	Markers          []string // files that triggered detection, e.g. "package.json"
}
