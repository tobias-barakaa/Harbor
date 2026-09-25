package docker

import (
	"fmt"
	"os"
	"path/filepath"

	"deployer/internal/app"
)

// EnsureDockerfile returns the path to a Dockerfile inside workDir,
// writing one if the project doesn't already ship its own. A
// project's own Dockerfile always wins — generation only fills the
// gap for runtimes detect/ already recognizes.
func EnsureDockerfile(workDir string, a app.Application) (string, error) {
	path := filepath.Join(workDir, "Dockerfile")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	content, err := generate(workDir, a)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func generate(workDir string, a app.Application) (string, error) {
	switch a.Runtime {
	case app.RuntimeNode:
		return nodeDockerfile(a), nil
	case app.RuntimePython:
		return pythonDockerfile(workDir, a), nil
	case app.RuntimeGo:
		return goDockerfile(a), nil
	default:
		return "", fmt.Errorf("no Dockerfile template for runtime %q — add your own Dockerfile to the project", a.Runtime)
	}
}

func nodeDockerfile(a app.Application) string {
	return fmt.Sprintf(`FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install --omit=dev
COPY . .
EXPOSE %d
CMD ["npm", "start"]
`, a.Port)
}

func pythonDockerfile(workDir string, a app.Application) string {
	deps := "COPY requirements.txt .\nRUN pip install --no-cache-dir -r requirements.txt"
	if !exists(filepath.Join(workDir, "requirements.txt")) {
		deps = "COPY pyproject.toml .\nRUN pip install --no-cache-dir ."
	}
	return fmt.Sprintf(`FROM python:3.12-slim
WORKDIR /app
%s
COPY . .
EXPOSE %d
%s
`, deps, a.Port, pythonCMD(workDir, a.Port))
}

// pythonCMD special-cases Django (manage.py runserver needs the bind
// address and port as explicit args, unlike a plain script) and
// otherwise falls back to the first recognized entry filename.
func pythonCMD(workDir string, port int) string {
	if exists(filepath.Join(workDir, "manage.py")) {
		return fmt.Sprintf(`CMD ["python", "manage.py", "runserver", "0.0.0.0:%d"]`, port)
	}
	for _, candidate := range []string{"app.py", "main.py", "run.py", "wsgi.py"} {
		if exists(filepath.Join(workDir, candidate)) {
			return fmt.Sprintf(`CMD ["python", "%s"]`, candidate)
		}
	}
	return `CMD ["python", "app.py"]` // best-effort default; fails loudly in container logs if wrong
}

// goDockerfile is a multi-stage build: the compiled binary is copied
// into a bare alpine image, so the final container doesn't carry the
// entire Go toolchain — just the ~10MB binary and a minimal base.
func goDockerfile(a app.Application) string {
	return fmt.Sprintf(`FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o /out/app .

FROM alpine:3.19
COPY --from=build /out/app /app
EXPOSE %d
CMD ["/app"]
`, a.Port)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}