package build

import (
	"bytes"
	"fmt"
	"path/filepath"

	"deployer/internal/app"
)

type pythonStrategy struct{}

func (pythonStrategy) Name() string { return "python" }

func (pythonStrategy) CanHandle(a app.Application) bool {
	return a.Runtime == app.RuntimePython
}

// venvPython/venvPip point at the executables inside the virtualenv
// Build creates, rather than whatever "python"/"pip" resolves to on
// PATH — that's the whole point of the venv: never touch system
// Python. This assumes a Linux venv layout (bin/), matching your
// Debian deployment target.
func venvPython(workDir string) string {
	return filepath.Join(workDir, ".venv", "bin", "python")
}

func venvPip(workDir string) string {
	return filepath.Join(workDir, ".venv", "bin", "pip")
}

func (pythonStrategy) Build(workDir string, a app.Application) (Result, error) {
	var logBuf bytes.Buffer

	// 1. Isolated environment first — nothing gets installed until
	// this exists, so a failed/missing python3-venv fails loudly here
	// rather than quietly installing into system site-packages.
	if err := runStep(workDir, &logBuf, "python3", "-m", "venv", ".venv"); err != nil {
		return Result{Success: false, Log: logBuf.String()}, err
	}

	// 2. Dependencies. requirements.txt is the common case; a
	// pyproject.toml with no requirements.txt is installed as the
	// project itself, which works whether it's setuptools- or
	// Poetry-backed as long as the backend is PEP 517 compliant.
	switch {
	case exists(filepath.Join(workDir, "requirements.txt")):
		if err := runStep(workDir, &logBuf, venvPip(workDir), "install", "-r", "requirements.txt"); err != nil {
			return Result{Success: false, Log: logBuf.String()}, err
		}
	case exists(filepath.Join(workDir, "pyproject.toml")):
		if err := runStep(workDir, &logBuf, venvPip(workDir), "install", "."); err != nil {
			return Result{Success: false, Log: logBuf.String()}, err
		}
	}

	startCmd, startArgs, err := pythonStartCommand(workDir, a)
	if err != nil {
		return Result{Success: false, Log: logBuf.String()}, err
	}

	return Result{
		Success:   true,
		StartCmd:  startCmd,
		StartArgs: startArgs,
		Log:       logBuf.String(),
	}, nil
}

// pythonStartCommand picks how to launch the app after install.
// Django's manage.py is special-cased — "python manage.py runserver"
// is the standard way to run it, and guessing at an entry filename
// would be wrong for it specifically. Everything else falls back to
// the first common entry filename found, run directly with the
// venv's python.
func pythonStartCommand(workDir string, a app.Application) (string, []string, error) {
	python := venvPython(workDir)

	if exists(filepath.Join(workDir, "manage.py")) {
		return python, []string{"manage.py", "runserver", fmt.Sprintf("0.0.0.0:%d", a.Port)}, nil
	}

	for _, candidate := range []string{"app.py", "main.py", "run.py", "wsgi.py"} {
		if exists(filepath.Join(workDir, candidate)) {
			return python, []string{candidate}, nil
		}
	}

	return "", nil, fmt.Errorf(
		"no recognizable Python entry point found (looked for manage.py, app.py, main.py, run.py, wsgi.py)")
}
