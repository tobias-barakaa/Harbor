package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".deployer")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "servers.json"), nil
}

type Store struct {
	servers map[string]Server
	path    string
}

func Open() (*Store, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	s := &Store{servers: map[string]Server{}, path: path}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	var list []Server
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	for _, srv := range list {
		s.servers[srv.Name] = srv
	}
	return s, nil
}

func (s *Store) Save() error {
	list := make([]Server, 0, len(s.servers))
	for _, srv := range s.servers {
		list = append(list, srv)
	}
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	// 0600: this file may contain a plaintext password (--password auth).
	return os.WriteFile(s.path, data, 0o600)
}

func (s *Store) Add(srv Server) error {
	if _, exists := s.servers[srv.Name]; exists {
		return fmt.Errorf("server %q already exists", srv.Name)
	}
	s.servers[srv.Name] = srv
	return s.Save()
}

func (s *Store) Get(name string) (Server, error) {
	srv, ok := s.servers[name]
	if !ok {
		return Server{}, fmt.Errorf("no server named %q", name)
	}
	return srv, nil
}

func (s *Store) List() []Server {
	list := make([]Server, 0, len(s.servers))
	for _, srv := range s.servers {
		list = append(list, srv)
	}
	return list
}

func (s *Store) Remove(name string) error {
	if _, ok := s.servers[name]; !ok {
		return fmt.Errorf("no server named %q", name)
	}
	delete(s.servers, name)
	return s.Save()
}
