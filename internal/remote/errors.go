package remote

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// classifyConnectErr rewrites a low-level dial/auth error into
// something that names the actual cause (connection refused, host
// unreachable, timeout, auth failure) instead of an opaque wrapped
// string. The original error is always preserved with %w — this adds
// context, it never hides anything.
func classifyConnectErr(err error, addr string) error {
	if err == nil {
		return nil
	}

	var netErr *net.OpError
	if errors.As(err, &netErr) {
		msg := netErr.Err.Error()
		switch {
		case strings.Contains(msg, "connection refused"):
			return fmt.Errorf("connection refused by %s (is sshd running, and the port correct?): %w", addr, err)
		case strings.Contains(msg, "no route to host"), strings.Contains(msg, "network is unreachable"):
			return fmt.Errorf("host %s unreachable: %w", addr, err)
		case netErr.Timeout():
			return fmt.Errorf("timed out reaching %s: %w", addr, err)
		}
	}

	if strings.Contains(err.Error(), "unable to authenticate") {
		return fmt.Errorf("SSH authentication failed for %s (check username/key/password): %w", addr, err)
	}

	return err
}