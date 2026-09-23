package main

import (
	"fmt"
	"os"

	"deployer/internal/inspect"
	"deployer/internal/report"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "inspect" {
		fmt.Println("Usage: go run ./cmd/deployer inspect <path-to-zip>")
		os.Exit(1)
	}

	apps, err := inspect.Zip(os.Args[2])
	if err != nil {
		fmt.Printf("Error inspecting zip file: %v\n", err)
		os.Exit(1)
	}

	report.Print(apps)
}