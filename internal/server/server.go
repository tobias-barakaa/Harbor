package server

type AuthMethod string

const (
	AuthPrivateKey AuthMethod = "private_key"
	AuthPassword   AuthMethod = "password"
)

// Server is everything needed to reach one remote machine. Nothing
// above this layer — build strategies, process lifecycle — needs to
// know or care whether it's a VPS, a Hetzner box, or your laptop.
type Server struct {
	Name           string     `json:"name"`
	Host           string     `json:"host"`
	Port           int        `json:"port"`
	Username       string     `json:"username"`
	AuthMethod     AuthMethod `json:"auth_method"`
	PrivateKeyPath string     `json:"private_key_path,omitempty"`
	Password       string     `json:"password,omitempty"` // see store.go — plaintext, flagged as tech debt
}
