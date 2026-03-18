package config

import (
	"os"
	"strings"
)

// Config holds all configuration for the flat tool
type Config struct {
	Verbose        bool
	NoBin          bool
	External       bool
	Exclude        []string
	IgnoreFile     string
	BypassChecksum bool
	JustAgents     bool
}

// LoadConfig loads configuration from environment variables and defaults
func LoadConfig() *Config {
	cfg := &Config{
		Verbose:    isEnvTrue("FLAT_VERBOSE"),
		IgnoreFile: ".flatignore",
	}
	return cfg
}

// isEnvTrue checks if an environment variable is set to true
func isEnvTrue(key string) bool {
	val := os.Getenv(key)
	return strings.ToLower(val) == "true" || val == "1"
}

// SetVerbose sets verbose mode
func (c *Config) SetVerbose(v bool) {
	c.Verbose = v
}

// SetNoBin sets binary skip mode
func (c *Config) SetNoBin(v bool) {
	c.NoBin = v
}

// SetExternal sets external reference mode
func (c *Config) SetExternal(v bool) {
	c.External = v
}

// SetExclude sets exclude patterns
func (c *Config) SetExclude(patterns []string) {
	c.Exclude = patterns
}

// SetIgnoreFile sets ignore file path
func (c *Config) SetIgnoreFile(path string) {
	c.IgnoreFile = path
}

// SetBypassChecksum sets checksum bypass flag
func (c *Config) SetBypassChecksum(v bool) {
	c.BypassChecksum = v
}

// SetJustAgents sets just-agents mode
func (c *Config) SetJustAgents(v bool) {
	c.JustAgents = v
}
