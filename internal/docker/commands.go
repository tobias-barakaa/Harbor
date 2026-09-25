package docker

import "fmt"

func BuildCommand(image, remoteDir string) string {
	return fmt.Sprintf("docker build -t %s %s", shellQuote(image), shellQuote(remoteDir))
}

// RemoveCommand stops and deletes any existing container with this
// name. It's written to never fail the calling shell (the `|| true`
// equivalent), so it's safe to run unconditionally before every
// deploy — including the very first one, when there's nothing to
// remove yet. That's what makes redeploying the same app idempotent.
func RemoveCommand(containerName string) string {
	return fmt.Sprintf("docker rm -f %s 2>/dev/null; true", shellQuote(containerName))
}

func RunCommand(containerName, image string, ports []PortMapping, volumes []Volume, env []EnvVar) string {
	cmd := fmt.Sprintf("docker run -d --name %s --restart unless-stopped", shellQuote(containerName))
	for _, p := range ports {
		cmd += " " + p.flag()
	}
	for _, v := range volumes {
		cmd += " " + v.flag()
	}
	for _, e := range env {
		cmd += " " + e.flag()
	}
	return cmd + " " + shellQuote(image)
}

func LogsCommand(containerName string, follow bool) string {
	if follow {
		return fmt.Sprintf("docker logs -f %s", shellQuote(containerName))
	}
	return fmt.Sprintf("docker logs %s", shellQuote(containerName))
}

// StatusCommand reports the container's state directly, rather than
// parsing `docker ps` table output — far less fragile if Docker
// changes its column formatting between versions.
func StatusCommand(containerName string) string {
	return fmt.Sprintf(`docker inspect -f '{{.State.Status}}' %s 2>/dev/null || echo "not found"`, shellQuote(containerName))
}

func StopCommand(containerName string) string  { return fmt.Sprintf("docker stop %s", shellQuote(containerName)) }
func StartCommand(containerName string) string { return fmt.Sprintf("docker start %s", shellQuote(containerName)) }