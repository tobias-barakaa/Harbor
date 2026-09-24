package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"deployer/internal/remote"
	"deployer/internal/server"
)

func cmdServer(args []string) {
	if len(args) < 1 {
		printServerUsage()
		os.Exit(1)
	}
	switch args[0] {
	case "add":
		cmdServerAdd(args[1:])
	case "list":
		cmdServerList()
	case "remove":
		cmdServerRemove(args[1:])
	case "test":
		cmdServerTest(args[1:])
	case "exec":
		cmdServerExec(args[1:])
	case "upload":
		cmdServerUpload(args[1:])
	default:
		printServerUsage()
		os.Exit(1)
	}
}

func printServerUsage() {
	fmt.Println(`Usage:
  deployer server add <name> --host <host> [--port 22] --user <user> (--key <path> | --password <pw>)
  deployer server list
  deployer server remove <name>
  deployer server test <name>
  deployer server exec <name> "<command>"
  deployer server upload <name> <local-path> <remote-dir>`)
}

func cmdServerAdd(args []string) {
	if len(args) < 1 {
		printServerUsage()
		os.Exit(1)
	}
	name, rest := args[0], args[1:]

	host := flagValue(rest, "--host")
	user := flagValue(rest, "--user")
	if host == "" || user == "" {
		fmt.Println("Error: --host and --user are required")
		os.Exit(1)
	}

	port := 22
	if p := flagValue(rest, "--port"); p != "" {
		parsed, err := strconv.Atoi(p)
		if err != nil {
			fmt.Println("Error: --port must be a number")
			os.Exit(1)
		}
		port = parsed
	}

	srv := server.Server{Name: name, Host: host, Port: port, Username: user}
	switch {
	case flagValue(rest, "--key") != "":
		srv.AuthMethod = server.AuthPrivateKey
		srv.PrivateKeyPath = flagValue(rest, "--key")
	case flagValue(rest, "--password") != "":
		fmt.Println("Warning: password auth is stored in plaintext in ~/.deployer/servers.json — prefer --key")
		srv.AuthMethod = server.AuthPassword
		srv.Password = flagValue(rest, "--password")
	default:
		fmt.Println("Error: specify either --key <path> or --password <pw>")
		os.Exit(1)
	}

	store, err := server.Open()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	if err := store.Add(srv); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Printf("Added server %q\n", name)
}

func cmdServerList() {
	store, err := server.Open()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	servers := store.List()
	if len(servers) == 0 {
		fmt.Println("No servers registered.")
		return
	}
	for _, s := range servers {
		fmt.Printf("%-15s %s@%s:%d (%s)\n", s.Name, s.Username, s.Host, s.Port, s.AuthMethod)
	}
}

func cmdServerRemove(args []string) {
	if len(args) < 1 {
		printServerUsage()
		os.Exit(1)
	}
	store, err := server.Open()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	if err := store.Remove(args[0]); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Printf("Removed server %q\n", args[0])
}

func cmdServerTest(args []string) {
	if len(args) < 1 {
		printServerUsage()
		os.Exit(1)
	}
	srv := loadServer(args[0])

	fmt.Printf("Testing server: %s\n\nHost: %s\nUser: %s\n\n", srv.Name, srv.Host, srv.Username)
	client, err := remote.Connect(srv)
	if err != nil {
		fmt.Println("✗ SSH connection failed:", err)
		os.Exit(1)
	}
	defer client.Close()

	if err := client.Test(); err != nil {
		fmt.Println("✗ Connected, but command execution failed:", err)
		os.Exit(1)
	}
	fmt.Println("✓ SSH connection successful")
}

func cmdServerExec(args []string) {
	if len(args) < 2 {
		printServerUsage()
		os.Exit(1)
	}
	srv := loadServer(args[0])
	command := args[1]

	client, err := remote.Connect(srv)
	if err != nil {
		fmt.Println("Error connecting:", err)
		os.Exit(1)
	}
	defer client.Close()

	result := client.Exec(command)
	fmt.Print(result.Stdout)
	if result.Stderr != "" {
		fmt.Fprint(os.Stderr, result.Stderr)
	}

	if result.Err != nil {
		fmt.Println("Error running command:", result.Err)
		os.Exit(1)
	}
	if result.ExitCode != 0 {
		// The command ran fine over SSH and just returned non-zero —
		// propagate that exit code rather than masking it as "an
		// error", so scripting against deployer (e.g. in CI) can
		// check $? meaningfully.
		os.Exit(result.ExitCode)
	}
}

func cmdServerUpload(args []string) {
	if len(args) < 3 {
		printServerUsage()
		os.Exit(1)
	}
	srv := loadServer(args[0])
	localPath, remoteDir := args[1], args[2]

	client, err := remote.Connect(srv)
	if err != nil {
		fmt.Println("Error connecting:", err)
		os.Exit(1)
	}
	defer client.Close()

	info, err := os.Stat(localPath)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Printf("Uploading %s to %s:%s...\n", localPath, srv.Name, remoteDir)
	if info.IsDir() {
		err = client.UploadDir(localPath, remoteDir)
	} else {
		// path.Join, not string concatenation — remoteDir is always a
		// Unix-style remote path, and Join correctly collapses a
		// trailing slash instead of producing "dir//file".
		err = client.UploadFile(localPath, path.Join(remoteDir, filepath.Base(localPath)))
	}
	if err != nil {
		fmt.Println("Upload failed:", err)
		os.Exit(1)
	}
	fmt.Println("✓ Upload complete")
}

func loadServer(name string) server.Server {
	store, err := server.Open()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	srv, err := store.Get(name)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	return srv
}