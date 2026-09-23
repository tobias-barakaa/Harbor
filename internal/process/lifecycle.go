package process

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// Run starts name+args detached inside workDir as the leader of its
// own process group (Setsid gives it a new session where pgid == pid),
// so the whole tree it spawns — npm, and whatever npm forks — can
// later be signaled together instead of just the top-level command.
func Run(appName, workDir, command string, args []string) error {
	dir, err := stateDir()
	if err != nil {
		return err
	}
	logPath := filepath.Join(dir, appName+".log")

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer logFile.Close()

	cmd := exec.Command(command, args...)
	cmd.Dir = workDir
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		return err
	}

	return saveState(State{
		Name:      appName,
		PID:       cmd.Process.Pid,
		WorkDir:   workDir,
		Command:   command,
		Args:      args,
		LogPath:   logPath,
		StartedAt: time.Now().Format(time.RFC3339),
	})
}

// Stop signals the app's entire process group: SIGTERM first, then
// polls for actual exit, then escalates to SIGKILL if the group
// hasn't died within the grace period. It blocks until the group is
// confirmed gone (or force-killed), so callers — including Restart —
// can rely on the app truly being down when Stop returns.
func Stop(appName string) error {
	s, err := loadState(appName)
	if err != nil {
		return fmt.Errorf("no known app named %q: %w", appName, err)
	}
	return stopProcessGroup(s.PID, 5*time.Second)
}

func stopProcessGroup(pid int, grace time.Duration) error {
	pgid, err := syscall.Getpgid(pid)
	if err != nil {
		if err == syscall.ESRCH {
			return nil // already gone — nothing to do
		}
		return err
	}

	if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil && err != syscall.ESRCH {
		return err
	}

	deadline := time.Now().Add(grace)
	for time.Now().Before(deadline) {
		if !groupAlive(pgid) {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Didn't exit cleanly within the grace period — force it.
	if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil && err != syscall.ESRCH {
		return err
	}
	return nil
}

// groupAlive checks for the group's existence with signal 0 (sends no
// actual signal). EPERM still means it exists, just owned differently;
// only ESRCH means it's actually gone.
func groupAlive(pgid int) bool {
	err := syscall.Kill(-pgid, syscall.Signal(0))
	return err == nil || err == syscall.EPERM
}

// Restart stops the app (blocking until it's actually dead) and runs
// it again with the same command it was originally started with. The
// short sleep after Stop is just to let the OS release the port —
// Stop itself already guarantees the processes are gone.
func Restart(appName string) error {
	s, err := loadState(appName)
	if err != nil {
		return fmt.Errorf("no known app named %q: %w", appName, err)
	}
	if err := Stop(appName); err != nil {
		return err
	}
	time.Sleep(300 * time.Millisecond)
	return Run(s.Name, s.WorkDir, s.Command, s.Args)
}

// Info reports whether the app's process group is currently alive,
// alongside the metadata recorded when it was started.
type Info struct {
	Name      string
	Running   bool
	PID       int
	WorkDir   string
	StartedAt string
}

func Status(appName string) (Info, error) {
	s, err := loadState(appName)
	if err != nil {
		return Info{}, fmt.Errorf("no known app named %q: %w", appName, err)
	}
	pgid, err := syscall.Getpgid(s.PID)
	running := err == nil && groupAlive(pgid)
	return Info{
		Name:      s.Name,
		Running:   running,
		PID:       s.PID,
		WorkDir:   s.WorkDir,
		StartedAt: s.StartedAt,
	}, nil
}

// Logs writes the app's log file to w. With follow=true it keeps
// reading as new output arrives, like `tail -f`.
func Logs(appName string, follow bool, w io.Writer) error {
	s, err := loadState(appName)
	if err != nil {
		return fmt.Errorf("no known app named %q: %w", appName, err)
	}
	f, err := os.Open(s.LogPath)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	io.Copy(w, reader)
	if !follow {
		return nil
	}
	for {
		time.Sleep(300 * time.Millisecond)
		io.Copy(w, reader)
	}
}