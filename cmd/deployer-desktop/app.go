package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"deployer/internal/app"
	"deployer/internal/catalog"
	"deployer/internal/deploy"
	"deployer/internal/inspect"
	"deployer/internal/remote"
	"deployer/internal/server"
)

type App struct {
	ctx context.Context

	// inFlight guards against duplicate concurrent operations on the
	// same named resource — e.g. clicking Deploy twice before the
	// first call returns. Keyed by "<kind>:<name>" so a deployment
	// and a server can never collide even sharing a name.
	inFlightMu sync.Mutex
	inFlight   map[string]bool

	// logStreams tracks each deployment's stop channel so
	// StopDeploymentLogs can find and close it.
	logStreamsMu sync.Mutex
	logStreams   map[string]chan struct{}
}

func NewApp() *App {
	return &App{inFlight: map[string]bool{}, logStreams: map[string]chan struct{}{}}
}

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

func (a *App) lock(key string) bool {
	a.inFlightMu.Lock()
	defer a.inFlightMu.Unlock()
	if a.inFlight[key] {
		return false
	}
	a.inFlight[key] = true
	return true
}

func (a *App) unlock(key string) {
	a.inFlightMu.Lock()
	defer a.inFlightMu.Unlock()
	delete(a.inFlight, key)
}

// ---------- Applications ----------

type ApplicationEntry struct {
	ID        string `json:"id"`
	Path      string `json:"path"`
	Name      string `json:"name"`
	Runtime   string `json:"runtime"`
	Framework string `json:"framework"`
	Strategy  string `json:"strategy"`
	Port      int    `json:"port"`
	AddedAt   string `json:"addedAt"`
}

func toApplicationEntry(e catalog.Entry) ApplicationEntry {
	return ApplicationEntry{
		ID: e.ID, Path: e.Path, Name: e.Name,
		Runtime: e.Runtime, Framework: e.Framework, Strategy: e.Strategy,
		Port: e.Port, AddedAt: e.AddedAt,
	}
}

func (a *App) PickDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "Select application directory"})
}

func (a *App) PickZipFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Select application zip",
		Filters: []runtime.FileFilter{{DisplayName: "Zip files (*.zip)", Pattern: "*.zip"}},
	})
}

// AddApplication runs detection on path and registers every detected
// application — a path can contain more than one (a monorepo's
// frontend and backend both get their own entry, sharing Path but
// distinguished by Name, exactly like inspect.Path already
// distinguishes them for the CLI).
func (a *App) AddApplication(path string) ([]ApplicationEntry, error) {
	if path == "" {
		return nil, validationError("no path given")
	}
	apps, err := inspect.Path(path)
	if err != nil {
		return nil, categorized(categoryInspection, err)
	}
	if len(apps) == 0 {
		return nil, categorized(categoryInspection, fmt.Errorf("no recognizable application found at %s", path))
	}

	store, err := catalog.Open()
	if err != nil {
		return nil, categorized(categoryUnexpected, err)
	}

	result := make([]ApplicationEntry, 0, len(apps))
	for _, ap := range apps {
		entry, err := store.Add(catalog.Entry{
			Path: path, Name: ap.Name, Runtime: string(ap.Runtime),
			Framework: string(ap.Framework), Strategy: string(ap.Strategy), Port: ap.Port,
		})
		if err != nil {
			return nil, categorized(categoryUnexpected, err)
		}
		result = append(result, toApplicationEntry(entry))
	}
	return result, nil
}

func (a *App) ListApplications() ([]ApplicationEntry, error) {
	store, err := catalog.Open()
	if err != nil {
		return nil, categorized(categoryUnexpected, err)
	}
	entries := store.List()
	result := make([]ApplicationEntry, 0, len(entries))
	for _, e := range entries {
		result = append(result, toApplicationEntry(e))
	}
	return result, nil
}

// ReinspectApplication re-runs detection against the entry's stored
// path and refreshes its cached metadata — useful after the user
// changes the project without re-adding it from scratch.
func (a *App) ReinspectApplication(id string) (ApplicationEntry, error) {
	store, err := catalog.Open()
	if err != nil {
		return ApplicationEntry{}, categorized(categoryUnexpected, err)
	}
	entry, err := store.Get(id)
	if err != nil {
		return ApplicationEntry{}, categorized(categoryValidation, err)
	}

	apps, err := inspect.Path(entry.Path)
	if err != nil {
		return ApplicationEntry{}, categorized(categoryInspection, err)
	}
	var match *app.Application
	for i := range apps {
		if apps[i].Name == entry.Name {
			match = &apps[i]
			break
		}
	}
	if match == nil {
		return ApplicationEntry{}, categorized(categoryInspection, fmt.Errorf("application %q no longer found at %s", entry.Name, entry.Path))
	}

	entry.Runtime = string(match.Runtime)
	entry.Framework = string(match.Framework)
	entry.Strategy = string(match.Strategy)
	entry.Port = match.Port
	if err := store.Update(entry); err != nil {
		return ApplicationEntry{}, categorized(categoryUnexpected, err)
	}
	return toApplicationEntry(entry), nil
}

func (a *App) RemoveApplication(id string) error {
	store, err := catalog.Open()
	if err != nil {
		return categorized(categoryUnexpected, err)
	}
	return categorized(categoryUnexpected, store.Remove(id))
}

// ---------- Servers ----------

type ServerInfo struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	AuthMethod string `json:"authMethod"`
}

func toServerInfo(s server.Server) ServerInfo {
	return ServerInfo{Name: s.Name, Host: s.Host, Port: s.Port, Username: s.Username, AuthMethod: string(s.AuthMethod)}
}

type AddServerInput struct {
	Name           string `json:"name"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	AuthMethod     string `json:"authMethod"`
	PrivateKeyPath string `json:"privateKeyPath"`
	Password       string `json:"password"`
}

func (a *App) ListServers() ([]ServerInfo, error) {
	store, err := server.Open()
	if err != nil {
		return nil, categorized(categoryUnexpected, err)
	}
	list := store.List()
	result := make([]ServerInfo, 0, len(list))
	for _, s := range list {
		result = append(result, toServerInfo(s))
	}
	return result, nil
}

func (a *App) AddServer(input AddServerInput) error {
	if input.Name == "" || input.Host == "" || input.Username == "" {
		return validationError("name, host, and username are required")
	}
	if input.AuthMethod != string(server.AuthPrivateKey) && input.AuthMethod != string(server.AuthPassword) {
		return validationError(`authMethod must be "private_key" or "password"`)
	}
	if input.AuthMethod == string(server.AuthPrivateKey) && input.PrivateKeyPath == "" {
		return validationError("privateKeyPath is required for private_key auth")
	}
	if input.AuthMethod == string(server.AuthPassword) && input.Password == "" {
		return validationError("password is required for password auth")
	}
	port := input.Port
	if port == 0 {
		port = 22
	}

	store, err := server.Open()
	if err != nil {
		return categorized(categoryUnexpected, err)
	}
	return categorized(categoryUnexpected, store.Add(server.Server{
		Name: input.Name, Host: input.Host, Port: port, Username: input.Username,
		AuthMethod:     server.AuthMethod(input.AuthMethod),
		PrivateKeyPath: input.PrivateKeyPath, Password: input.Password,
	}))
}

func (a *App) TestServer(name string) error {
	key := "server:" + name
	if !a.lock(key) {
		return validationError("a test for this server is already in progress")
	}
	defer a.unlock(key)

	store, err := server.Open()
	if err != nil {
		return categorized(categoryUnexpected, err)
	}
	srv, err := store.Get(name)
	if err != nil {
		return categorized(categoryValidation, err)
	}
	return categorized(categoryConnection, remote.TestConnection(srv))
}

func (a *App) RemoveServer(name string) error {
	store, err := server.Open()
	if err != nil {
		return categorized(categoryUnexpected, err)
	}
	return categorized(categoryUnexpected, store.Remove(name))
}

// ---------- Deployments ----------

type PortMappingInfo struct {
	Host      int `json:"host"`
	Container int `json:"container"`
}

type DeploymentInfo struct {
	Name          string            `json:"name"`
	ServerName    string            `json:"serverName"`
	ContainerName string            `json:"containerName"`
	Image         string            `json:"image"`
	Ports         []PortMappingInfo `json:"ports"`
	DeployedAt    string            `json:"deployedAt"`
	Status        string            `json:"status"`
}

func toDeploymentInfo(d deploy.Deployment, status string) DeploymentInfo {
	ports := make([]PortMappingInfo, 0, len(d.Ports))
	for _, p := range d.Ports {
		ports = append(ports, PortMappingInfo{Host: p.Host, Container: p.Container})
	}
	return DeploymentInfo{
		Name: d.Name, ServerName: d.ServerName, ContainerName: d.ContainerName,
		Image: d.Image, Ports: ports, DeployedAt: d.DeployedAt, Status: status,
	}
}

// normalizeDockerStatus collapses Docker's own container states into
// the three the UI can act on. Docker distinguishes more
// (created/restarting/paused/dead/...); the desktop only needs to
// know whether Stop makes sense (running), Restart makes sense
// (stopped), or neither (not_found). Pending/Building/Deploying from
// the original spec are deliberately NOT here — they exist only
// client-side, for the duration of a Deploy() call itself, since
// deploy.Run has no way to report intermediate phases back to the
// caller. A real Building-vs-Deploying split would need deploy.Run
// itself to report phase progress — a genuine backend addition, out
// of scope for this pass, flagged rather than faked.
func normalizeDockerStatus(raw string) string {
	switch raw {
	case "running":
		return "running"
	case "not found", "":
		return "not_found"
	default:
		return "stopped"
	}
}

type DeployInput struct {
	ApplicationID string `json:"applicationId"`
	ServerName    string `json:"serverName"`
}

func (a *App) ListDeployments() ([]DeploymentInfo, error) {
	store, err := deploy.Open()
	if err != nil {
		return nil, categorized(categoryUnexpected, err)
	}
	list := store.List()
	result := make([]DeploymentInfo, 0, len(list))
	for _, d := range list {
		raw, _ := deploy.Status(d.Name) // best-effort — a failed live status check shouldn't hide the deployment from the list
		result = append(result, toDeploymentInfo(d, normalizeDockerStatus(raw)))
	}
	return result, nil
}

func (a *App) Deploy(input DeployInput) (DeploymentInfo, error) {
	if input.ApplicationID == "" || input.ServerName == "" {
		return DeploymentInfo{}, validationError("applicationId and serverName are required")
	}

	catStore, err := catalog.Open()
	if err != nil {
		return DeploymentInfo{}, categorized(categoryUnexpected, err)
	}
	entry, err := catStore.Get(input.ApplicationID)
	if err != nil {
		return DeploymentInfo{}, categorized(categoryValidation, err)
	}

	key := "deploy:" + entry.Name
	if !a.lock(key) {
		return DeploymentInfo{}, validationError("a deployment operation for this application is already in progress")
	}
	defer a.unlock(key)

	d, err := deploy.Run(deploy.Spec{Target: entry.Path, AppName: entry.Name, ServerName: input.ServerName})
	if err != nil {
		return DeploymentInfo{}, classifyDeployError(err)
	}

	raw, _ := deploy.Status(d.Name)
	return toDeploymentInfo(d, normalizeDockerStatus(raw)), nil
}

func (a *App) GetDeploymentStatus(name string) (DeploymentInfo, error) {
	store, err := deploy.Open()
	if err != nil {
		return DeploymentInfo{}, categorized(categoryUnexpected, err)
	}
	d, err := store.Get(name)
	if err != nil {
		return DeploymentInfo{}, categorized(categoryValidation, err)
	}
	raw, err := deploy.Status(name)
	if err != nil {
		return DeploymentInfo{}, categorized(categoryConnection, err)
	}
	return toDeploymentInfo(d, normalizeDockerStatus(raw)), nil
}

func (a *App) StopDeployment(name string) error {
	key := "deploy:" + name
	if !a.lock(key) {
		return validationError("a deployment operation for this application is already in progress")
	}
	defer a.unlock(key)
	return categorized(categoryDeployment, deploy.Stop(name))
}

func (a *App) RestartDeployment(name string) error {
	key := "deploy:" + name
	if !a.lock(key) {
		return validationError("a deployment operation for this application is already in progress")
	}
	defer a.unlock(key)
	return categorized(categoryDeployment, deploy.Restart(name))
}

func (a *App) RemoveDeployment(name string) error {
	key := "deploy:" + name
	if !a.lock(key) {
		return validationError("a deployment operation for this application is already in progress")
	}
	defer a.unlock(key)
	return categorized(categoryDeployment, deploy.Remove(name))
}

// ---------- Logs ----------

// StreamDeploymentLogs starts following a deployment's container logs
// in the background and emits each line as a Wails runtime event —
// "deployment-log" with {name, line} — so the frontend renders them
// live instead of calling a blocking Go method from a click handler.
// A "deployment-log-closed" event with {name, error} fires when the
// stream ends, whether the container stopped logging or
// StopDeploymentLogs was called.
func (a *App) StreamDeploymentLogs(name string) error {
	a.logStreamsMu.Lock()
	if _, exists := a.logStreams[name]; exists {
		a.logStreamsMu.Unlock()
		return validationError("already streaming logs for this deployment")
	}
	stop := make(chan struct{})
	a.logStreams[name] = stop
	a.logStreamsMu.Unlock()

	pr, pw := io.Pipe()

	go func() {
		err := deploy.StreamLogs(name, pw, stop)
		pw.CloseWithError(err)
	}()

	go func() {
		scanner := bufio.NewScanner(pr)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			runtime.EventsEmit(a.ctx, "deployment-log", map[string]string{
				"name": name,
				"line": scanner.Text(),
			})
		}

		a.logStreamsMu.Lock()
		delete(a.logStreams, name)
		a.logStreamsMu.Unlock()

		errMsg := ""
		if err := scanner.Err(); err != nil && err != io.EOF {
			errMsg = err.Error()
		}
		runtime.EventsEmit(a.ctx, "deployment-log-closed", map[string]string{
			"name":  name,
			"error": errMsg,
		})
	}()

	return nil
}

func (a *App) StopDeploymentLogs(name string) error {
	a.logStreamsMu.Lock()
	stop, exists := a.logStreams[name]
	a.logStreamsMu.Unlock()
	if !exists {
		return nil // already stopped — idempotent, matches Stop semantics elsewhere in this codebase
	}
	close(stop)
	return nil
}