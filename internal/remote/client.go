package remote

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"

	"deployer/internal/server"
)

// Client wraps one SSH connection to one Server.
type Client struct {
	conn *ssh.Client
}

// Connect dials the server and authenticates with whichever method
// its config specifies.
//
// Host key checking is intentionally permissive right now
// (InsecureIgnoreHostKey) — acceptable while you're only pointing
// this at boxes you already control, but this is the first thing to
// harden before Harbor talks to anything you don't fully trust: as-is
// it can't detect a machine-in-the-middle on the connection.
func Connect(srv server.Server) (*Client, error) {
	authMethod, err := authMethodFor(srv)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User:            srv.Username,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", srv.Host, srv.Port)
	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, classifyConnectErr(fmt.Errorf("connecting to %s: %w", srv.Name, err), addr)
	}
	return &Client{conn: conn}, nil
}

func authMethodFor(srv server.Server) (ssh.AuthMethod, error) {
	switch srv.AuthMethod {
	case server.AuthPrivateKey:
		key, err := os.ReadFile(srv.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("reading private key %s: %w", srv.PrivateKeyPath, err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("parsing private key %s: %w", srv.PrivateKeyPath, err)
		}
		return ssh.PublicKeys(signer), nil
	case server.AuthPassword:
		return ssh.Password(srv.Password), nil
	default:
		return nil, fmt.Errorf("unknown auth method %q", srv.AuthMethod)
	}
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// Test runs a trivial command to confirm the connection actually
// works end-to-end — a successful Dial only proves the handshake
// worked, not that a session can run commands.
func (c *Client) Test() error {
	result := c.Exec("echo ok")
	if result.Err != nil {
		return result.Err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("test command exited with status %d", result.ExitCode)
	}
	return nil
}