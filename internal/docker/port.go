package docker

import "deployer/internal/app"

// ContainerPort returns the port the application listens on inside
// the container. Application.Port always represents the external
// port users reach the application on.
func ContainerPort(a app.Application) int {
	if a.Strategy == app.StrategyStatic {
		return 80
	}

	return a.Port
}