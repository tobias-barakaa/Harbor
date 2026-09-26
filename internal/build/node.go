package build

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"deployer/internal/app"
)

type nodeStrategy struct{}

func (nodeStrategy) Name() string { return "node" }

func (nodeStrategy) CanHandle(a app.Application) bool {
	return a.Runtime == app.RuntimeNode
}

type packageJSON struct {
	Scripts map[string]string `json:"scripts"`
}

func (nodeStrategy) Build(workDir string, a app.Application) (Result, error) {
	var logBuf bytes.Buffer

	installCmd := packageManagerInstallCmd(workDir)
	if err := runStep(workDir, &logBuf, installCmd[0], installCmd[1:]...); err != nil {
		return Result{Success: false, Log: logBuf.String()}, err
	}

	buildCmd := a.BuildCmd
	if buildCmd == "" {
		// No framework-specific build command — fall back to
		// package.json's own "build" script, same behavior as before
		// framework detection existed.
		if pkg, err := readPackageJSON(workDir); err == nil {
			if _, hasBuild := pkg.Scripts["build"]; hasBuild {
				buildCmd = "npm run build"
			}
		}
	}
	if buildCmd != "" {
		if err := runStep(workDir, &logBuf, "sh", "-c", buildCmd); err != nil {
			return Result{Success: false, Log: logBuf.String()}, err
		}
	}

	if a.Strategy == app.StrategyStatic {
		// Nothing to "start" — serve the build output. `npx serve`
		// is the standard zero-config static file server, so there's
		// no separate install step needed for it.
		outputDir := a.OutputDir
		if outputDir == "" {
			outputDir = "dist"
		}
		return Result{
			Success:   true,
			StartCmd:  "npx",
			StartArgs: []string{"serve", outputDir, "-l", strconv.Itoa(a.Port)},
			Log:       logBuf.String(),
		}, nil
	}

	startCmd, startArgs := "npm", []string{"start"}
	if pkg, err := readPackageJSON(workDir); err == nil {
		if _, hasStart := pkg.Scripts["start"]; !hasStart {
			startCmd, startArgs = "node", []string{entryFile(workDir)}
		}
	}

	return Result{
		Success:   true,
		StartCmd:  startCmd,
		StartArgs: startArgs,
		Log:       logBuf.String(),
	}, nil
}

func readPackageJSON(workDir string) (packageJSON, error) {
	data, err := os.ReadFile(filepath.Join(workDir, "package.json"))
	if err != nil {
		return packageJSON{}, err
	}
	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return packageJSON{}, err
	}
	return pkg, nil
}

func packageManagerInstallCmd(workDir string) []string {
	if exists(filepath.Join(workDir, "pnpm-lock.yaml")) {
		return []string{"pnpm", "install"}
	}
	if exists(filepath.Join(workDir, "yarn.lock")) {
		return []string{"yarn", "install"}
	}
	return []string{"npm", "install"}
}

func entryFile(workDir string) string {
	for _, candidate := range []string{"index.js", "server.js", "app.js", "main.js"} {
		if exists(filepath.Join(workDir, candidate)) {
			return candidate
		}
	}
	return "index.js"
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func runStep(workDir string, log *bytes.Buffer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = workDir
	cmd.Stdout = log
	cmd.Stderr = log
	return cmd.Run()
}