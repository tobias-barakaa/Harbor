package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"deployer/internal/docker"
)

// Deployment records what a `deployer deploy` created, so later
// status/logs/stop/restart/remove calls don't need the ports, server
// name, or anything else repeated on the command line.
type Deployment struct {
	Name          string               `json:"name"`
	ServerName    string               `json:"server_name"`
	ContainerName string               `json:"container_name"`
	Image         string               `json:"image"`
	RemoteDir     string               `json:"remote_dir"`
	Ports         []docker.PortMapping `json:"ports"`
	Volumes       []docker.Volume      `json:"volumes"`
	Env           []docker.EnvVar      `json:"env"`
	DeployedAt    string               `json:"deployed_at"`
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".deployer")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "deployments.json"), nil
}

type Store struct {
	deployments map[string]Deployment
	path        string
}

func Open() (*Store, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	s := &Store{deployments: map[string]Deployment{}, path: path}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	var list []Deployment
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	for _, d := range list {
		s.deployments[d.Name] = d
	}
	return s, nil
}

func (s *Store) Save() error {
	list := make([]Deployment, 0, len(s.deployments))
	for _, d := range s.deployments {
		list = append(list, d)
	}
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *Store) Put(d Deployment) error {
	s.deployments[d.Name] = d
	return s.Save()
}

func (s *Store) Get(name string) (Deployment, error) {
	d, ok := s.deployments[name]
	if !ok {
		return Deployment{}, fmt.Errorf("no deployment named %q", name)
	}
	return d, nil
}

func (s *Store) List() []Deployment {
	list := make([]Deployment, 0, len(s.deployments))
	for _, d := range s.deployments {
		list = append(list, d)
	}
	return list
}

func (s *Store) Remove(name string) error {
	if _, ok := s.deployments[name]; !ok {
		return fmt.Errorf("no deployment named %q", name)
	}
	delete(s.deployments, name)
	return s.Save()
}