package process

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// State is everything needed to find, stop, or restart a running app
// after the original `deployer run` process has exited. It's the
// bridge between separate CLI invocations.
type State struct {
	Name      string   `json:"name"`
	PID       int      `json:"pid"`
	WorkDir   string   `json:"work_dir"`
	Command   string   `json:"command"`
	Args      []string `json:"args"`
	LogPath   string   `json:"log_path"`
	StartedAt string   `json:"started_at"`
}

func stateDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".deployer", "state")
	return dir, os.MkdirAll(dir, 0o755)
}

func statePath(name string) (string, error) {
	dir, err := stateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".json"), nil
}

func saveState(s State) error {
	path, err := statePath(s.Name)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func loadState(name string) (State, error) {
	path, err := statePath(name)
	if err != nil {
		return State{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return State{}, err
	}
	var s State
	err = json.Unmarshal(data, &s)
	return s, err
}
