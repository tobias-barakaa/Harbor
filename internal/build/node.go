package build

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"deployer/internal/app"
)

type nodeStrategy struct{}

func (nodeStrategy) Name() string { return "node" }

func (nodeStrategy) CanHandle(a app.Application) bool {
	return a.Runtime == app.RuntimeNode
}

// packageJSON is the subset of package.json we actually need: which
// scripts exist, so we know whether there's a build step and how to
// start the app afterward.
type packageJSON struct {
	Scripts map[string]string `json:"scripts"`
}

func (nodeStrategy) Build(workDir string, a app.Application) (Result, error) {
	var logBuf bytes.Buffer

	pkg, err := readPackageJSON(workDir)
	if err != nil {
		return Result{}, err
	}

	installCmd := packageManagerInstallCmd(workDir)
	if err := runStep(workDir, &logBuf, installCmd[0], installCmd[1:]...); err != nil {
		return Result{Success: false, Log: logBuf.String()}, err
	}

	if _, hasBuild := pkg.Scripts["build"]; hasBuild {
		if err := runStep(workDir, &logBuf, "npm", "run", "build"); err != nil {
			return Result{Success: false, Log: logBuf.String()}, err
		}
	}

	startCmd, startArgs := "npm", []string{"start"}
	if _, hasStart := pkg.Scripts["start"]; !hasStart {
		// No "start" script — fall back to running the entry file
		// directly rather than a "npm start" that's guaranteed to fail.
		startCmd, startArgs = "node", []string{entryFile(workDir)}
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

// packageManagerInstallCmd picks npm/pnpm/yarn based on which lockfile
// is present, since that's the strongest signal of what the project
// actually expects — installing with the wrong one can produce a
// different dependency tree than what was tested.
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
	return "index.js" // best-effort default; will fail loudly if wrong, which is fine
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
