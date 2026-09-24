package remote

import (
	"bytes"

	"golang.org/x/crypto/ssh"
)

// ExecResult is a remote command's full outcome. ExitCode is only
// meaningful when Err is nil — a non-nil Err means the command never
// ran to completion at all (session couldn't open, connection died
// mid-command), which is a different failure mode than "ran and
// returned non-zero." Callers must check both, not just Err.
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

// Exec runs cmd in a fresh SSH session (sessions are single-use — one
// command per session) and reports the real outcome, distinguishing a
// non-zero exit from a connection-level failure instead of collapsing
// both into one generic error.
func (c *Client) Exec(cmd string) ExecResult {
	session, err := c.conn.NewSession()
	if err != nil {
		return ExecResult{Err: err}
	}
	defer session.Close()

	var outBuf, errBuf bytes.Buffer
	session.Stdout = &outBuf
	session.Stderr = &errBuf

	runErr := session.Run(cmd)
	result := ExecResult{Stdout: outBuf.String(), Stderr: errBuf.String()}

	if runErr == nil {
		return result
	}

	if exitErr, ok := runErr.(*ssh.ExitError); ok {
		// The command ran and returned non-zero — not a connection
		// failure, so Err stays nil and the caller decides what to do
		// with ExitCode.
		result.ExitCode = exitErr.ExitStatus()
		return result
	}

	// Anything else (session dropped, command never ran) is a real
	// error, not "the command failed."
	result.Err = runErr
	return result
}