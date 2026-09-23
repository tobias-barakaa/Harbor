package main

import (
	"fmt"
	"os"

	"deployer/internal/inspect"
	"deployer/internal/report"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "inspect" {
		fmt.Println("Usage: go run ./cmd/deployer inspect <path-to-zip-or-directory>")
		os.Exit(1)
	}

	apps, err := inspect.Path(os.Args[2])
	if err != nil {
		fmt.Printf("Error inspecting %s: %v\n", os.Args[2], err)
		os.Exit(1)
	}

	report.Print(apps)
}
