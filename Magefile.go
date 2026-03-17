//go:build mage

package mage

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	projectName = "flat"
	version     = "0.1.0"
)

// Build builds the flat binary
func Build() error {
	mg.Deps(prepare)
	fmt.Println("Building flat...")

	cmd := exec.Command("go", "build", "-o", projectName, "-ldflags", fmt.Sprintf("-X main.version=%s", version))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("✓ Built %s successfully\n", projectName)
	return nil
}

// BuildDebug builds the flat binary with debug symbols
func BuildDebug() error {
	mg.Deps(prepare)
	fmt.Println("Building flat (debug)...")

	cmd := exec.Command("go", "build", "-o", projectName, "-gcflags", "all=-N -l")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("✓ Built %s (debug) successfully\n", projectName)
	return nil
}

// BuildRelease builds the flat binary with optimizations
func BuildRelease() error {
	mg.Deps(prepare)
	fmt.Println("Building flat (release)...")

	cmd := exec.Command("go", "build", "-o", projectName, "-ldflags", fmt.Sprintf("-s -w -X main.version=%s", version))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("✓ Built %s (release) successfully\n", projectName)
	return nil
}

// Install installs the flat binary to $GOPATH/bin
func Install() error {
	mg.Deps(Build)
	fmt.Println("Installing flat to $GOPATH/bin...")

	cmd := exec.Command("go", "install", "-ldflags", fmt.Sprintf("-X main.version=%s", version))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	// Get the GOPATH
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME"), "go")
	}

	binDir := filepath.Join(gopath, "bin")
	binaryPath := filepath.Join(binDir, projectName)

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Try to copy to GOPATH/bin
		src, err := filepath.Abs(projectName)
		if err != nil {
			return err
		}

		if err := os.Rename(src, binaryPath); err != nil {
			return fmt.Errorf("failed to copy to %s: %w", binaryPath, err)
		}
	}

	fmt.Printf("✓ Installed to %s\n", binaryPath)
	return nil
}

// Uninstall uninstalls the flat binary
func Uninstall() error {
	fmt.Println("Uninstalling flat...")

	binDir := filepath.Join(os.Getenv("HOME"), ".flat")
	binaryPath := filepath.Join(binDir, projectName)

	if err := os.Remove(binaryPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("uninstall failed: %w", err)
	}

	fmt.Println("✓ Uninstalled successfully")
	return nil
}

// Clean removes build artifacts
func Clean() error {
	fmt.Println("Cleaning build artifacts...")

	// Remove binary
	if err := os.Remove(projectName); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clean failed: %w", err)
	}

	// Remove build directory
	if err := os.RemoveAll("build"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clean failed: %w", err)
	}

	fmt.Println("✓ Cleaned successfully")
	return nil
}

// Test runs all tests
func Test() error {
	mg.Deps(prepare)
	fmt.Println("Running tests...")

	cmd := exec.Command("go", "test", "./...", "-v", "-cover")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("test failed: %w", err)
	}

	fmt.Println("✓ Tests passed")
	return nil
}

// TestShort runs tests without verbose output
func TestShort() error {
	mg.Deps(prepare)
	fmt.Println("Running tests...")

	cmd := exec.Command("go", "test", "./...", "-cover")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("test failed: %w", err)
	}

	fmt.Println("✓ Tests passed")
	return nil
}

// TestRace runs tests with race detector
func TestRace() error {
	mg.Deps(prepare)
	fmt.Println("Running tests with race detector...")

	cmd := exec.Command("go", "test", "./...", "-race", "-v")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("test failed: %w", err)
	}

	fmt.Println("✓ Race tests passed")
	return nil
}

// Coverage runs tests with coverage report
func Coverage() error {
	mg.Deps(prepare)
	fmt.Println("Running tests with coverage...")

	cmd := exec.Command("go", "test", "./...", "-coverprofile=coverage.out", "-covermode=count")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("test failed: %w", err)
	}

	fmt.Println("✓ Coverage report generated (coverage.out)")
	return nil
}

// CoverageHTML generates HTML coverage report
func CoverageHTML() error {
	mg.Deps(Coverage)
	fmt.Println("Generating HTML coverage report...")

	cmd := exec.Command("go", "tool", "cover", "-html=coverage.out", "-o=coverage.html")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("coverage html failed: %w", err)
	}

	fmt.Println("✓ HTML coverage report generated (coverage.html)")
	return nil
}

// Lint runs linters
func Lint() error {
	mg.Deps(prepare)
	fmt.Println("Running linters...")

	// Run go vet
	if err := sh.Run("go", "vet", "./..."); err != nil {
		return fmt.Errorf("go vet failed: %w", err)
	}

	// Run gofmt check
	if err := sh.Run("gofmt", "-l", "."); err != nil {
		return fmt.Errorf("gofmt found issues: %w", err)
	}

	fmt.Println("✓ Linting passed")
	return nil
}

// Format runs gofmt
func Format() error {
	fmt.Println("Formatting code...")

	cmd := exec.Command("gofmt", "-w", ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("format failed: %w", err)
	}

	fmt.Println("✓ Code formatted")
	return nil
}

// Vet runs go vet
func Vet() error {
	mg.Deps(prepare)
	fmt.Println("Running go vet...")

	cmd := exec.Command("go", "vet", "./...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("vet failed: %w", err)
	}

	fmt.Println("✓ Go vet passed")
	return nil
}

// Tidy runs go mod tidy
func Tidy() error {
	fmt.Println("Running go mod tidy...")

	cmd := exec.Command("go", "mod", "tidy")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tidy failed: %w", err)
	}

	fmt.Println("✓ Dependencies tidied")
	return nil
}

// Prepare runs pre-build checks
func prepare() {
	fmt.Println("Preparing build...")
}

// CrossBuild builds for multiple platforms
func CrossBuild() error {
	mg.Deps(prepare)
	fmt.Println("Building for multiple platforms...")

	platforms := []string{
		"linux/amd64",
		"linux/arm64",
		"darwin/amd64",
		"darwin/arm64",
		"windows/amd64",
	}

	for _, platform := range platforms {
		parts := filepath.SplitList(platform)
		goos := parts[0]
		goarch := parts[1]

		fmt.Printf("Building for %s/%s...\n", goos, goarch)

		outFile := fmt.Sprintf("%s-%s-%s", projectName, goos, goarch)
		if goos == "windows" {
			outFile += ".exe"
		}

		cmd := exec.Command("go", "build", "-o", outFile, "-ldflags", fmt.Sprintf("-X main.version=%s", version))
		cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("build for %s/%s failed: %w", goos, goarch, err)
		}

		fmt.Printf("✓ Built %s\n", outFile)
	}

	fmt.Println("✓ Cross-build complete")
	return nil
}

// Docs builds documentation (placeholder)
func Docs() error {
	fmt.Println("Documentation is in docs/ directory")
	fmt.Println("Available docs:")
	fmt.Println("  - docs/phase-0.md")
	fmt.Println("  - docs/phase-1.md")
	fmt.Println("  - docs/phase-2.md")
	fmt.Println("  - docs/phase-3.md")
	fmt.Println("  - docs/overview.md")
	fmt.Println("  - docs/checklist.md")
	fmt.Println("  - docs/test-coverage.md")
	fmt.Println("  - README.md")
	return nil
}

// Help displays available commands
func Help() {
	fmt.Println("Flat Build Commands")
	fmt.Println("===================")
	fmt.Println()
	fmt.Println("Build:")
	fmt.Println("  mage build          Build flat binary")
	fmt.Println("  mage builddebug     Build with debug symbols")
	fmt.Println("  mage buildrelease   Build optimized release binary")
	fmt.Println("  mage crossbuild     Build for multiple platforms")
	fmt.Println()
	fmt.Println("Install:")
	fmt.Println("  mage install        Install to $GOPATH/bin")
	fmt.Println("  mage uninstall      Uninstall from $GOPATH/bin")
	fmt.Println()
	fmt.Println("Test:")
	fmt.Println("  mage test           Run all tests")
	fmt.Println("  mage testshort      Run tests without verbose output")
	fmt.Println("  mage testrace       Run tests with race detector")
	fmt.Println("  mage coverage       Run tests with coverage report")
	fmt.Println("  mage coveragehtml   Generate HTML coverage report")
	fmt.Println()
	fmt.Println("Lint:")
	fmt.Println("  mage lint           Run linters (go vet, gofmt)")
	fmt.Println("  mage format         Format code with gofmt")
	fmt.Println("  mage vet            Run go vet")
	fmt.Println()
	fmt.Println("Clean:")
	fmt.Println("  mage clean          Remove build artifacts")
	fmt.Println()
	fmt.Println("Other:")
	fmt.Println("  mage tidy           Run go mod tidy")
	fmt.Println("  mage docs           Display documentation info")
	fmt.Println("  mage help           Display this help")
}
