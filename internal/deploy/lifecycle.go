package deploy

import (
	"fmt"
	"io"
	"strings"

	"deployer/internal/docker"
	"deployer/internal/remote"
	"deployer/internal/server"
)

func connectFor(d Deployment) (*remote.Client, error) {
	srvStore, err := server.Open()
	if err != nil {
		return nil, err
	}
	srv, err := srvStore.Get(d.ServerName)
	if err != nil {
		return nil, err
	}
	return remote.Connect(srv)
}

func Status(name string) (string, error) {
	store, err := Open()
	if err != nil {
		return "", err
	}
	d, err := store.Get(name)
	if err != nil {
		return "", err
	}
	client, err := connectFor(d)
	if err != nil {
		return "", err
	}
	defer client.Close()

	result := client.Exec(docker.StatusCommand(d.ContainerName))
	if result.Err != nil {
		return "", result.Err
	}
	return strings.TrimSpace(result.Stdout), nil
}

// Logs streams to w. With follow=true, the remote `docker logs -f`
// keeps the SSH session open and pushes new lines as they're
// written — this call blocks until the caller kills it or the
// container stops producing logs.
func Logs(name string, follow bool, w io.Writer) error {
	store, err := Open()
	if err != nil {
		return err
	}
	d, err := store.Get(name)
	if err != nil {
		return err
	}
	client, err := connectFor(d)
	if err != nil {
		return err
	}
	defer client.Close()

	result := client.Exec(docker.LogsCommand(d.ContainerName, follow))
	fmt.Fprint(w, result.Stdout)
	if result.Stderr != "" {
		fmt.Fprint(w, result.Stderr)
	}
	return result.Err
}

func Stop(name string) error {
	store, err := Open()
	if err != nil {
		return err
	}
	d, err := store.Get(name)
	if err != nil {
		return err
	}
	client, err := connectFor(d)
	if err != nil {
		return err
	}
	defer client.Close()
	return runOrFail(client, docker.StopCommand(d.ContainerName))
}

func Restart(name string) error {
	store, err := Open()
	if err != nil {
		return err
	}
	d, err := store.Get(name)
	if err != nil {
		return err
	}
	client, err := connectFor(d)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := runOrFail(client, docker.StopCommand(d.ContainerName)); err != nil {
		return err
	}
	return runOrFail(client, docker.StartCommand(d.ContainerName))
}

// Remove stops and deletes the container remotely, then forgets the
// local record entirely — not just a stop. A later `deployer deploy`
// of the same app builds and runs it from scratch.
func Remove(name string) error {
	store, err := Open()
	if err != nil {
		return err
	}
	d, err := store.Get(name)
	if err != nil {
		return err
	}
	client, err := connectFor(d)
	if err != nil {
		return err
	}
	defer client.Close()

	client.Exec(docker.RemoveCommand(d.ContainerName))
	return store.Remove(name)
}