package config

import (
	"os"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear any environment variables
	os.Setenv("FLAT_VERBOSE", "")

	cfg := LoadConfig()

	if cfg.Verbose != false {
		t.Errorf("Verbose should be false by default, got %v", cfg.Verbose)
	}
	if cfg.IgnoreFile != ".flatignore" {
		t.Errorf("IgnoreFile should be '.flatignore' by default, got %q", cfg.IgnoreFile)
	}
}

func TestLoadConfig_VerboseFromEnv(t *testing.T) {
	testCases := []struct {
		name   string
		value  string
		should bool
	}{
		{"true", "true", true},
		{"True", "True", true},
		{"TRUE", "TRUE", true},
		{"1", "1", true},
		{"false", "false", false},
		{"False", "False", false},
		{"0", "0", false},
		{"empty", "", false},
		{"undefined", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.value != "" {
				os.Setenv("FLAT_VERBOSE", tc.value)
			} else {
				os.Unsetenv("FLAT_VERBOSE")
			}

			cfg := LoadConfig()
			if cfg.Verbose != tc.should {
				t.Errorf("Verbose should be %v when FLAT_VERBOSE=%q, got %v", tc.should, tc.value, cfg.Verbose)
			}
		})
	}
}

func TestConfig_Setters(t *testing.T) {
	cfg := &Config{}

	// Test SetVerbose
	cfg.SetVerbose(true)
	if !cfg.Verbose {
		t.Error("SetVerbose(true) did not set Verbose to true")
	}

	cfg.SetVerbose(false)
	if cfg.Verbose {
		t.Error("SetVerbose(false) did not set Verbose to false")
	}

	// Test SetNoBin
	cfg.SetNoBin(true)
	if !cfg.NoBin {
		t.Error("SetNoBin(true) did not set NoBin to true")
	}

	cfg.SetNoBin(false)
	if cfg.NoBin {
		t.Error("SetNoBin(false) did not set NoBin to false")
	}

	// Test SetExternal
	cfg.SetExternal(true)
	if !cfg.External {
		t.Error("SetExternal(true) did not set External to true")
	}

	cfg.SetExternal(false)
	if cfg.External {
		t.Error("SetExternal(false) did not set External to false")
	}

	// Test SetExclude
	cfg.SetExclude([]string{"*.bin", "*.exe"})
	if len(cfg.Exclude) != 2 {
		t.Errorf("SetExclude did not set Exclude correctly, got %d patterns", len(cfg.Exclude))
	}
	if cfg.Exclude[0] != "*.bin" {
		t.Errorf("SetExclude first pattern is %q, expected '*.bin'", cfg.Exclude[0])
	}

	// Test SetIgnoreFile
	cfg.SetIgnoreFile(".customignore")
	if cfg.IgnoreFile != ".customignore" {
		t.Errorf("SetIgnoreFile did not set IgnoreFile correctly, got %q", cfg.IgnoreFile)
	}

	// Test SetBypassChecksum
	cfg.SetBypassChecksum(true)
	if !cfg.BypassChecksum {
		t.Error("SetBypassChecksum(true) did not set BypassChecksum to true")
	}

	cfg.SetBypassChecksum(false)
	if cfg.BypassChecksum {
		t.Error("SetBypassChecksum(false) did not set BypassChecksum to false")
	}
}

func TestConfig_Isolation(t *testing.T) {
	// Test that LoadConfig creates independent instances
	os.Setenv("FLAT_VERBOSE", "true")

	cfg1 := LoadConfig()
	os.Setenv("FLAT_VERBOSE", "false")

	cfg2 := LoadConfig()

	if cfg1.Verbose != true {
		t.Error("cfg1 should have Verbose=true")
	}
	if cfg2.Verbose != false {
		t.Error("cfg2 should have Verbose=false")
	}
}

func TestConfig_EmptyConfig(t *testing.T) {
	cfg := &Config{}

	// Verify all fields have zero values
	if cfg.Verbose != false {
		t.Error("Config.Verbose should be false")
	}
	if cfg.NoBin != false {
		t.Error("Config.NoBin should be false")
	}
	if cfg.External != false {
		t.Error("Config.External should be false")
	}
	if cfg.Exclude != nil && len(cfg.Exclude) != 0 {
		t.Error("Config.Exclude should be nil or empty")
	}
	if cfg.IgnoreFile != "" {
		t.Error("Config.IgnoreFile should be empty")
	}
	if cfg.BypassChecksum != false {
		t.Error("Config.BypassChecksum should be false")
	}
}

func TestConfig_ExcludePatterns(t *testing.T) {
	cfg := &Config{}

	// Empty slice
	cfg.SetExclude(nil)
	if cfg.Exclude != nil {
		t.Error("SetExclude(nil) should set Exclude to nil")
	}

	// Empty slice literal
	cfg.SetExclude([]string{})
	if len(cfg.Exclude) != 0 {
		t.Error("SetExclude([]string{}) should set Exclude to empty slice")
	}

	// Multiple patterns
	patterns := []string{"*.bin", "*.exe", "*.so"}
	cfg.SetExclude(patterns)
	if len(cfg.Exclude) != len(patterns) {
		t.Error("SetExclude should preserve all patterns")
	}
}
