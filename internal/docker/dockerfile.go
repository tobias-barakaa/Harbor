package docker

import (
	"fmt"
	"os"
	"path/filepath"

	"deployer/internal/app"
)

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
	if a.Strategy == app.StrategyStatic {
		return staticNodeDockerfile(a)
	}
	return fmt.Sprintf(`FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install --omit=dev
COPY . .
EXPOSE %d
CMD ["npm", "start"]
`, a.Port)
}

// staticNodeDockerfile builds in a full Node stage (the framework's
// build command needs devDependencies) and serves ONLY the output
// directory from a separate, much smaller nginx stage — the final
// image never carries node_modules or the framework's toolchain.
// The container always listens on 80 (nginx's default) regardless of
// a.Port — a.Port is the public-facing port this gets mapped to
// (handled by deploy.Run), not anything inside the container itself.
func staticNodeDockerfile(a app.Application) string {
	outputDir := a.OutputDir
	if outputDir == "" {
		outputDir = "dist"
	}
	buildCmd := a.BuildCmd
	if buildCmd == "" {
		buildCmd = "npm run build"
	}
	return fmt.Sprintf(`FROM node:20-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN %s

FROM nginx:alpine
COPY --from=build /app/%s /usr/share/nginx/html
EXPOSE 80
`, buildCmd, outputDir)
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

func pythonCMD(workDir string, port int) string {
	if exists(filepath.Join(workDir, "manage.py")) {
		return fmt.Sprintf(`CMD ["python", "manage.py", "runserver", "0.0.0.0:%d"]`, port)
	}
	for _, candidate := range []string{"app.py", "main.py", "run.py", "wsgi.py"} {
		if exists(filepath.Join(workDir, candidate)) {
			return fmt.Sprintf(`CMD ["python", "%s"]`, candidate)
		}
	}
	return `CMD ["python", "app.py"]`
}

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