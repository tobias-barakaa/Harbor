package deploy

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"deployer/internal/app"
	"deployer/internal/docker"
	"deployer/internal/inspect"
	"deployer/internal/remote"
	"deployer/internal/server"
	"deployer/internal/workspace"
)

// Spec is everything the caller supplies. Application detection and
// the Dockerfile fill in everything else — the caller never needs to
// know or state the runtime.
type Spec struct {
	Target     string // local zip or directory
	AppName    string // --app, when Target has more than one app
	ServerName string
	Ports      []docker.PortMapping
	Volumes    []docker.Volume
	Env        []docker.EnvVar
}

// Run executes one full deploy: resolve the app, materialize it on
// disk, generate a Dockerfile if needed, upload, build the image
// remotely, and (re)start the container. Running this again for the
// same app is a redeploy — RemoveCommand makes replacing the old
// container safe even on the very first run.
func Run(spec Spec) (Deployment, error) {
	apps, err := inspect.Path(spec.Target)
	if err != nil {
		return Deployment{}, err
	}
	a, err := pickApp(apps, spec.Target, spec.AppName)
	if err != nil {
		return Deployment{}, err
	}

	workDir, cleanup, err := workspace.Prepare(spec.Target, a)
	if err != nil {
		return Deployment{}, fmt.Errorf("preparing workspace: %w", err)
	}
	defer cleanup()

	if _, err := docker.EnsureDockerfile(workDir, a); err != nil {
		return Deployment{}, err
	}

	srvStore, err := server.Open()
	if err != nil {
		return Deployment{}, err
	}
	srv, err := srvStore.Get(spec.ServerName)
	if err != nil {
		return Deployment{}, err
	}

	client, err := remote.Connect(srv)
	if err != nil {
		return Deployment{}, err
	}
	defer client.Close()

	if err := runOrFail(client, "docker version"); err != nil {
		return Deployment{}, fmt.Errorf("docker not available on %s (is it installed?): %w", srv.Name, err)
	}

	containerName := sanitizeName(a.Name)
	image := containerName + ":latest"
	remoteDir := path.Join("/opt/deployer/apps", containerName)

	if err := client.UploadDir(workDir, remoteDir); err != nil {
		return Deployment{}, fmt.Errorf("uploading to %s: %w", srv.Name, err)
	}

	if err := runOrFail(client, docker.BuildCommand(image, remoteDir)); err != nil {
		return Deployment{}, fmt.Errorf("docker build: %w", err)
	}

	client.Exec(docker.RemoveCommand(containerName)) // idempotent — ignore result

	ports := spec.Ports
	if len(ports) == 0 {
		ports = []docker.PortMapping{{Host: a.Port, Container: a.Port}}
	}

	if err := runOrFail(client, docker.RunCommand(containerName, image, ports, spec.Volumes, spec.Env)); err != nil {
		return Deployment{}, fmt.Errorf("docker run: %w", err)
	}

	d := Deployment{
		Name:          a.Name,
		ServerName:    srv.Name,
		ContainerName: containerName,
		Image:         image,
		RemoteDir:     remoteDir,
		Ports:         ports,
		Volumes:       spec.Volumes,
		Env:           spec.Env,
		DeployedAt:    time.Now().Format(time.RFC3339),
	}

	store, err := Open()
	if err != nil {
		return d, err // the deployment itself succeeded — only the local record failed
	}
	return d, store.Put(d)
}

func pickApp(apps []app.Application, target, name string) (app.Application, error) {
	if name != "" {
		for _, a := range apps {
			if a.Name == name {
				return a, nil
			}
		}
		return app.Application{}, fmt.Errorf("no app named %q found in %s", name, target)
	}
	if len(apps) == 1 {
		return apps[0], nil
	}
	return app.Application{}, fmt.Errorf("multiple apps found in %s — specify one with --app", target)
}

var nameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9_.-]`)

// sanitizeName makes a detected app name safe as both a Docker
// container name and image tag — both only permit [a-zA-Z0-9_.-].
// Note: this means two different apps that sanitize to the same name
// (e.g. "My App!" and "my-app") would collide on one server — not
// handled yet, flagged rather than silently wrong.
func sanitizeName(name string) string {
	return strings.ToLower(nameSanitizer.ReplaceAllString(name, "-"))
}

func runOrFail(client *remote.Client, cmd string) error {
	result := client.Exec(cmd)
	if result.Err != nil {
		return result.Err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("exit %d: %s", result.ExitCode, strings.TrimSpace(result.Stderr))
	}
	return nil
}