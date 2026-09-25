package docker

import (
	"fmt"
	"strconv"
	"strings"
)

type PortMapping struct {
	Host      int `json:"host"`
	Container int `json:"container"`
}

func (p PortMapping) flag() string {
	return fmt.Sprintf("-p %d:%d", p.Host, p.Container)
}

// ParsePort accepts "8080:3000" (host:container) or just "3000",
// which maps the same port on both sides — the common case when you
// just want the container's own port exposed as-is.
func ParsePort(s string) (PortMapping, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 1 {
		p, err := strconv.Atoi(parts[0])
		if err != nil {
			return PortMapping{}, fmt.Errorf("invalid port %q: %w", s, err)
		}
		return PortMapping{Host: p, Container: p}, nil
	}
	host, err := strconv.Atoi(parts[0])
	if err != nil {
		return PortMapping{}, fmt.Errorf("invalid host port in %q: %w", s, err)
	}
	container, err := strconv.Atoi(parts[1])
	if err != nil {
		return PortMapping{}, fmt.Errorf("invalid container port in %q: %w", s, err)
	}
	return PortMapping{Host: host, Container: container}, nil
}

// Volume is one -v flag. Source is either a host path or a named
// Docker volume — both forms are passed straight through, since
// docker itself tells them apart (a leading "/" or "." means a bind
// mount; anything else is treated as a named volume).
type Volume struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

func (v Volume) flag() string {
	return fmt.Sprintf("-v %s:%s", v.Source, v.Target)
}

func ParseVolume(s string) (Volume, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return Volume{}, fmt.Errorf("invalid volume %q — expected source:target", s)
	}
	return Volume{Source: parts[0], Target: parts[1]}, nil
}

type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (e EnvVar) flag() string {
	// Value is shell-quoted because it's interpolated into a remote
	// command string (see run.go) — without this, a value containing
	// a space or shell metacharacter could break the command, or
	// worse, inject an extra one.
	return fmt.Sprintf("-e %s=%s", e.Key, shellQuote(e.Value))
}

func ParseEnv(s string) (EnvVar, error) {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) != 2 {
		return EnvVar{}, fmt.Errorf("invalid env var %q — expected KEY=VALUE", s)
	}
	return EnvVar{Key: parts[0], Value: parts[1]}, nil
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}