//go:build !mage
// +build !mage

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"flat/config"
	"flat/version"
)

var (
	Verbose    bool
	IgnoreFile string
)

// Run executes the flat command
func Run() error {
	cfg := &config.Config{
		Verbose:    Verbose,
		IgnoreFile: IgnoreFile,
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
			flattencmd := FlattenCmd()
			flattencmd.Cfg = cfg
			if err := flattencmd.Execute([]string{cwd, fmdxPath}); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return nil
		}

		fmt.Printf("Error: %s already exists. Use 'flat flatten' or 'flat unflatten'\n", filepath.Base(cwd)+".fmdx")
		os.Exit(1)
	}

	switch args[0] {
	case "flatten":
		// Need at least 3 args (flatten + source-dir + output.fmdx)
		// Additional args can be flags
		if len(args) < 3 {
			fmt.Println("Usage: flat flatten <source-dir> <output.fmdx>")
			fmt.Println("Flags:")
			fmt.Println("  -v, --verbose          verbose output")
			fmt.Println("      --no-bin           skip binary files")
			fmt.Println("      --external         external file references")
			fmt.Println("      --exclude strings  exclude patterns")
			fmt.Println("      --ignore-file      ignore file path (default \".flatignore\")")
			return nil
		}

		flattencmd := FlattenCmd()
		flattencmd.Cfg = cfg
		if err := flattencmd.Execute(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "unflatten":
		// Need at least 3 args (unflatten + input.fmdx + dest_dir)
		// Additional args can be flags
		if len(args) < 3 {
			fmt.Println("Usage: flat unflatten <input.fmdx> <destination-dir>")
			fmt.Println("Flags:")
			fmt.Println("  -v, --verbose           verbose output")
			fmt.Println("      --bypass-checksum   skip checksum verification")
			return nil
		}

		unflattencmd := UnflattenCmd()
		unflattencmd.Cfg = cfg
		if err := unflattencmd.Execute(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "version":
		fmt.Printf("flat version %s\n", version.Version)
		fmt.Printf("Commit: %s\n", version.Commit)
		fmt.Printf("Built: %s\n", version.Date)

	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		fmt.Println("Available commands: flatten, unflatten, version")
	}

	return nil
}
