package detect

import "deployer/internal/app"

// frameworkProfile captures what's true about a specific framework
// that a generic "it's Node.js" guess gets wrong: its own build
// command, its own output location, and — the thing that started
// this whole change — its own conventional port.
type frameworkProfile struct {
	Marker    string
	Framework app.Framework
	Strategy  app.Strategy
	BuildCmd  string
	OutputDir string
	Port      int
}

// nodeFrameworks is checked in order; the first marker file found
// wins. Port values match each framework's own documented default
// (Astro's preview server is 4321, Vite's is 4173, Next's dev/start
// default is 3000) — not a made-up convention.
var nodeFrameworks = []frameworkProfile{
	{Marker: "astro.config.mjs", Framework: app.FrameworkAstro, Strategy: app.StrategyStatic, BuildCmd: "npm run build", OutputDir: "dist", Port: 4321},
	{Marker: "astro.config.ts", Framework: app.FrameworkAstro, Strategy: app.StrategyStatic, BuildCmd: "npm run build", OutputDir: "dist", Port: 4321},
	{Marker: "astro.config.js", Framework: app.FrameworkAstro, Strategy: app.StrategyStatic, BuildCmd: "npm run build", OutputDir: "dist", Port: 4321},
	{Marker: "next.config.js", Framework: app.FrameworkNext, Strategy: app.StrategyServer, BuildCmd: "npm run build", Port: 3000},
	{Marker: "next.config.mjs", Framework: app.FrameworkNext, Strategy: app.StrategyServer, BuildCmd: "npm run build", Port: 3000},
	{Marker: "vite.config.js", Framework: app.FrameworkVite, Strategy: app.StrategyStatic, BuildCmd: "npm run build", OutputDir: "dist", Port: 4173},
	{Marker: "vite.config.ts", Framework: app.FrameworkVite, Strategy: app.StrategyStatic, BuildCmd: "npm run build", OutputDir: "dist", Port: 4173},
}

// detectNodeFramework looks for a known framework's marker file
// alongside package.json. ok=false means a plain Node.js project —
// the caller falls back to generic Node.js defaults in that case.
func detectNodeFramework(files map[string]bool) (frameworkProfile, bool) {
	for _, fw := range nodeFrameworks {
		if files[fw.Marker] {
			return fw, true
		}
	}
	return frameworkProfile{}, false
}