package main

import (
	"fmt"
	"os"

	"deployer/internal/deploy"
	"deployer/internal/docker"
)

func cmdDeploy(args []string) {
	if len(args) < 1 {
		printDeployUsage()
		os.Exit(1)
	}
	switch args[0] {
	case "status":
		cmdDeployStatus(args[1:])
	case "logs":
		cmdDeployLogs(args[1:])
	case "stop":
		cmdDeployStop(args[1:])
	case "restart":
		cmdDeployRestart(args[1:])
	case "remove":
		cmdDeployRemove(args[1:])
	case "list":
		cmdDeployList()
	default:
		cmdDeployRun(args) // args[0] is a target path, not a subcommand
	}
}

func printDeployUsage() {
	fmt.Println(`Usage:
  deployer deploy <path-to-zip-or-directory> --server <name> [--app name]
                   [--port host:container]... [--volume source:target]... [--env KEY=VAL]...
  deployer deploy status  <name>
  deployer deploy logs    <name> [--follow]
  deployer deploy stop    <name>
  deployer deploy restart <name>
  deployer deploy remove  <name>
  deployer deploy list`)
}

// flagValues collects every occurrence of a repeatable flag, unlike
// flagValue (in main.go) which only returns the first.
func flagValues(args []string, flag string) []string {
	var values []string
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			values = append(values, args[i+1])
		}
	}
	return values
}

func cmdDeployRun(args []string) {
	target, rest := args[0], args[1:]

	serverName := flagValue(rest, "--server")
	if serverName == "" {
		fmt.Println("Error: --server is required")
		os.Exit(1)
	}

	spec := deploy.Spec{Target: target, AppName: flagValue(rest, "--app"), ServerName: serverName}

	for _, p := range flagValues(rest, "--port") {
		pm, err := docker.ParsePort(p)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		spec.Ports = append(spec.Ports, pm)
	}
	for _, v := range flagValues(rest, "--volume") {
		vol, err := docker.ParseVolume(v)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		spec.Volumes = append(spec.Volumes, vol)
	}
	for _, e := range flagValues(rest, "--env") {
		ev, err := docker.ParseEnv(e)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		spec.Env = append(spec.Env, ev)
	}

	fmt.Printf("Deploying %s to %s...\n", target, serverName)
	d, err := deploy.Run(spec)
	if err != nil {
		fmt.Println("Deploy failed:", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Deployed %q as container %q\n", d.Name, d.ContainerName)
	for _, p := range d.Ports {
		fmt.Printf("  port: %d -> %d\n", p.Host, p.Container)
	}
}

func cmdDeployStatus(args []string) {
	if len(args) < 1 {
		printDeployUsage()
		os.Exit(1)
	}
	status, err := deploy.Status(args[0])
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Printf("%s: %s\n", args[0], status)
}

func cmdDeployLogs(args []string) {
	if len(args) < 1 {
		printDeployUsage()
		os.Exit(1)
	}
	follow := contains(args[1:], "--follow")
	if err := deploy.Logs(args[0], follow, os.Stdout); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func cmdDeployStop(args []string) {
	if len(args) < 1 {
		printDeployUsage()
		os.Exit(1)
	}
	if err := deploy.Stop(args[0]); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Printf("Stopped %q\n", args[0])
}

func cmdDeployRestart(args []string) {
	if len(args) < 1 {
		printDeployUsage()
		os.Exit(1)
	}
	if err := deploy.Restart(args[0]); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Printf("Restarted %q\n", args[0])
}

func cmdDeployRemove(args []string) {
	if len(args) < 1 {
		printDeployUsage()
		os.Exit(1)
	}
	if err := deploy.Remove(args[0]); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Printf("Removed %q\n", args[0])
}

func cmdDeployList() {
	store, err := deploy.Open()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	deployments := store.List()
	if len(deployments) == 0 {
		fmt.Println("No deployments.")
		return
	}
	for _, d := range deployments {
		fmt.Printf("%-15s server=%s container=%s image=%s\n", d.Name, d.ServerName, d.ContainerName, d.Image)
	}
}