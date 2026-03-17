package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	version = "0.1.0"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	cfg := &Config{
		Verbose:    isEnvTrue("FLAT_VERBOSE"),
		IgnoreFile: ".flatignore",
	}

	args := os.Args[1:]

	if len(args) == 0 {
		// Default mode: check for .fmdx and auto-flatten if not present
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmdxPath := filepath.Join(cwd, filepath.Base(cwd)+".fmdx")

		if _, err := os.Stat(fmdxPath); os.IsNotExist(err) {
			fmt.Printf("Auto-flattening %s to %s\n", cwd, fmdxPath)
			// TODO: Implement flatten
			fmt.Println("Flatten command not implemented yet")
			return
		}

		fmt.Printf("Error: %s already exists. Use 'flat flatten' or 'flat unflatten'\n", filepath.Base(cwd)+".fmdx")
		os.Exit(1)
	}

	switch args[0] {
	case "flatten":
		if len(args) != 3 {
			fmt.Println("Usage: flat flatten <source-dir> <output.fmdx>")
			return
		}
		fmt.Printf("Flatten command not implemented yet\n")
		fmt.Printf("Source: %s\n", args[1])
		fmt.Printf("Output: %s\n", args[2])
		fmt.Printf("Verbose: %v\n", cfg.Verbose)

	case "unflatten":
		if len(args) != 3 {
			fmt.Println("Usage: flat unflatten <input.fmdx> <destination-dir>")
			return
		}
		fmt.Printf("Unflatten command not implemented yet\n")
		fmt.Printf("Input: %s\n", args[1])
		fmt.Printf("Destination: %s\n", args[2])
		fmt.Printf("Verbose: %v\n", cfg.Verbose)

	case "version":
		printVersion()

	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		fmt.Println("Available commands: flatten, unflatten, version")
	}
}

func isEnvTrue(key string) bool {
	val := os.Getenv(key)
	return val == "true" || val == "1"
}

func printVersion() {
	fmt.Printf("flat version %s\n", version)
	fmt.Printf("Commit: %s\n", commit)
	fmt.Printf("Built: %s\n", date)
}

type Config struct {
	Verbose        bool
	NoBin          bool
	External       bool
	Exclude        []string
	IgnoreFile     string
	BypassChecksum bool
}
