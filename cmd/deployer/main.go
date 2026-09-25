package main

import (
	"fmt"
	"os"

	"deployer/internal/app"
	"deployer/internal/build"
	"deployer/internal/inspect"
	"deployer/internal/process"
	"deployer/internal/report"
	"deployer/internal/workspace"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "inspect":
		cmdInspect()
	case "build":
		cmdBuild()
	case "run":
		cmdRun()
	case "status":
		cmdStatus()
	case "logs":
		cmdLogs()
	case "stop":
		cmdStop()
	case "restart":
		cmdRestart()
	case "server":
		cmdServer(os.Args[2:])
	case "deploy":
		cmdDeploy(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage:
  deployer inspect <path-to-zip-or-directory>
  deployer build   <path-to-zip-or-directory> [--app name]
  deployer run     <path-to-zip-or-directory> [--app name]
  deployer status  <app-name>
  deployer logs    <app-name> [--follow]
  deployer stop    <app-name>
  deployer restart <app-name>
  deployer server  add|list|remove|test|exec|upload ...
  deployer deploy  <path-to-zip-or-directory> --server <name> [--app name] [--port ...] [--volume ...] [--env ...]
  deployer deploy  status|logs|stop|restart|remove|list ...`)
}

func resolveApp(target string, rest []string) (app.Application, error) {
	apps, err := inspect.Path(target)
	if err != nil {
		return app.Application{}, err
	}
	name := flagValue(rest, "--app")
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

func flagValue(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func contains(args []string, v string) bool {
	for _, a := range args {
		if a == v {
			return true
		}
	}
	return false
}

func cmdInspect() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}
	apps, err := inspect.Path(os.Args[2])
	if err != nil {
		fmt.Printf("Error inspecting %s: %v\n", os.Args[2], err)
		os.Exit(1)
	}
	report.Print(apps)
}

func cmdBuild() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}
	target := os.Args[2]
	a, err := resolveApp(target, os.Args[3:])
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	strategy, ok := build.For(a)
	if !ok {
		fmt.Printf("No build strategy for runtime %q yet\n", a.Runtime)
		os.Exit(1)
	}

	workDir, cleanup, err := workspace.Prepare(target, a)
	if err != nil {
		fmt.Println("Error preparing workspace:", err)
		os.Exit(1)
	}
	defer cleanup()

	result, err := strategy.Build(workDir, a)
	fmt.Print(result.Log)
	if err != nil {
		fmt.Println("Build failed:", err)
		os.Exit(1)
	}
	fmt.Printf("Build succeeded — start with: %s %v\n", result.StartCmd, result.StartArgs)
}

func cmdRun() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}
	target := os.Args[2]
	a, err := resolveApp(target, os.Args[3:])
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	strategy, ok := build.For(a)
	if !ok {
		fmt.Printf("No build strategy for runtime %q yet\n", a.Runtime)
		os.Exit(1)
	}

	workDir, _, err := workspace.Prepare(target, a)
	if err != nil {
		fmt.Println("Error preparing workspace:", err)
		os.Exit(1)
	}

	result, err := strategy.Build(workDir, a)
	fmt.Print(result.Log)
	if err != nil {
		fmt.Println("Build failed:", err)
		os.Exit(1)
	}

	if err := process.Run(a.Name, workDir, result.StartCmd, result.StartArgs); err != nil {
		fmt.Println("Failed to start:", err)
		os.Exit(1)
	}
	fmt.Printf("Started %q\n", a.Name)
}

func cmdStatus() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}
	info, err := process.Status(os.Args[2])
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	state := "stopped"
	if info.Running {
		state = "running"
	}
	fmt.Printf("%s: %s\n  pid:      %d\n  work dir: %s\n  started:  %s\n",
		info.Name, state, info.PID, info.WorkDir, info.StartedAt)
}

func cmdLogs() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}
	follow := contains(os.Args[3:], "--follow")
	if err := process.Logs(os.Args[2], follow, os.Stdout); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func cmdStop() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}
	if err := process.Stop(os.Args[2]); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Printf("Stopped %q\n", os.Args[2])
}

func cmdRestart() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}
	if err := process.Restart(os.Args[2]); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Printf("Restarted %q\n", os.Args[2])
}