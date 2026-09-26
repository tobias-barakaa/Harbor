package report

import (
	"fmt"

	"deployer/internal/app"
)

func Print(apps []app.Application) {
	if len(apps) == 0 {
		fmt.Println("No recognizable applications found.")
		return
	}
	for i, a := range apps {
		if i > 0 {
			fmt.Println()
		}
		fmt.Println("Application")
		fmt.Println("────────────")
		fmt.Printf("Name: %s\n\n", a.Name)

		fmt.Println("Detected:")
		for _, m := range a.Markers {
			fmt.Printf("  ✓ %s\n", m)
		}

		fmt.Printf("\nRuntime:\n  %s\n", a.Runtime)
		if a.Framework != "" {
			fmt.Printf("\nFramework:\n  %s\n", a.Framework)
		}
		fmt.Printf("\nStrategy:\n  %s\n", a.Strategy)
		if a.BuildCmd != "" {
			fmt.Printf("\nBuild:\n  %s\n", a.BuildCmd)
		}
		if a.Strategy == app.StrategyStatic {
			outputDir := a.OutputDir
			if outputDir == "" {
				outputDir = "dist"
			}
			fmt.Printf("\nOutput:\n  %s/\n", outputDir)
		}
		fmt.Printf("\nPossible port:\n  %d\n", a.Port)
		fmt.Printf("\nDeployment method:\n  %s\n", a.DeploymentMethod)
	}
}