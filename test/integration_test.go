package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"flat/cmd"
	"flat/config"
)

func TestIntegration_FlattenUnflatten_SingleFile(t *testing.T) {
	sourceDir := t.TempDir()
	restoreDir := t.TempDir()

	sourceFile := filepath.Join(sourceDir, "test.txt")
	content := "test content for integration"
	if err := os.WriteFile(sourceFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	outputDir := t.TempDir()
	fmdxFile := filepath.Join(outputDir, "output.fmdx")

	cfg := &config.Config{
		Verbose:        false,
		BypassChecksum: true,
	}

	flattencmd := cmd.FlattenCmd()
	flattencmd.Cfg = cfg
	if err := flattencmd.Execute([]string{sourceDir, fmdxFile}); err != nil {
		t.Fatalf("Flatten failed: %v", err)
	}

	unflattencmd := cmd.UnflattenCmd()
	unflattencmd.Cfg = cfg
	if err := unflattencmd.Execute([]string{fmdxFile, restoreDir}); err != nil {
		t.Fatalf("Unflatten failed: %v", err)
	}

	restoredFile := filepath.Join(restoreDir, "test.txt")
	restoredContent, err := os.ReadFile(restoredFile)
	if err != nil {
		t.Fatalf("Read restored file failed: %v", err)
	}

	if strings.TrimSpace(string(restoredContent)) != content {
		t.Errorf("Content mismatch: expected %q, got %q", content, string(restoredContent))
	}
}

func TestIntegration_FlattenUnflatten_Directory(t *testing.T) {
	sourceDir := t.TempDir()
	restoreDir := t.TempDir()

	projectDir := filepath.Join(sourceDir, "project")
	os.MkdirAll(filepath.Join(projectDir, "src"), 0755)
	if err := os.WriteFile(filepath.Join(projectDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "src", "utils.go"), []byte("package src"), 0644); err != nil {
		t.Fatalf("Failed to write utils.go: %v", err)
	}

	outputDir := t.TempDir()
	fmdxFile := filepath.Join(outputDir, "project.fmdx")

	cfg := &config.Config{
		Verbose:        false,
		BypassChecksum: true,
	}

	flattencmd := cmd.FlattenCmd()
	flattencmd.Cfg = cfg
	if err := flattencmd.Execute([]string{sourceDir, fmdxFile}); err != nil {
		t.Fatalf("Flatten failed: %v", err)
	}

	unflattencmd := cmd.UnflattenCmd()
	unflattencmd.Cfg = cfg
	if err := unflattencmd.Execute([]string{fmdxFile, restoreDir}); err != nil {
		t.Fatalf("Unflatten failed: %v", err)
	}

	// Check that at least one file was restored
	files, err := filepath.Glob(filepath.Join(restoreDir, "project", "*.go"))
	if err != nil {
		t.Fatalf("Failed to glob restored files: %v", err)
	}
	if len(files) == 0 {
		t.Error("No Go files restored")
	}
}

func TestIntegration_FlattenUnflatten_EmptyFile(t *testing.T) {
	sourceDir := t.TempDir()
	restoreDir := t.TempDir()

	emptyFile := filepath.Join(sourceDir, "empty.txt")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write empty file: %v", err)
	}

	outputDir := t.TempDir()
	fmdxFile := filepath.Join(outputDir, "output.fmdx")

	cfg := &config.Config{
		Verbose:        false,
		BypassChecksum: true,
	}

	flattencmd := cmd.FlattenCmd()
	flattencmd.Cfg = cfg
	if err := flattencmd.Execute([]string{sourceDir, fmdxFile}); err != nil {
		t.Fatalf("Flatten failed: %v", err)
	}

	unflattencmd := cmd.UnflattenCmd()
	unflattencmd.Cfg = cfg
	if err := unflattencmd.Execute([]string{fmdxFile, restoreDir}); err != nil {
		t.Fatalf("Unflatten failed: %v", err)
	}

	restoredFile := filepath.Join(restoreDir, "empty.txt")
	restoredContent, err := os.ReadFile(restoredFile)
	if err != nil {
		t.Fatalf("Read restored file failed: %v", err)
	}

	if len(restoredContent) != 0 {
		t.Errorf("Empty file should have 0 bytes, got %d", len(restoredContent))
	}
}
