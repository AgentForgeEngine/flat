//go:build !mage
// +build !mage

package cmd

import (
	"fmt"
	"os"

	"flat/config"
	"flat/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfg *config.Config
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "flat",
	Short: "Flat - Flatten directory trees into a single file",
	Long: `flat is a tool for flattening directory trees into a single .fmdx file.
It can also unflatten .fmdx files back into directory structures.`,
	Version: version.Version,
}

func init() {
	cobra.OnInitialize(initConfig)

	if cfg == nil {
		cfg = &config.Config{}
	}

	RootCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "verbose output")
	RootCmd.PersistentFlags().StringVarP(&cfg.IgnoreFile, "ignore-file", "", ".flatignore", "ignore file path")

	RootCmd.AddCommand(FlattenCmd())
	RootCmd.AddCommand(UnflattenCmd())
	RootCmd.AddCommand(VersionCmd())
}

func initConfig() {
	viper.SetConfigName(".flat")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.flat")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		return
	}

	if cfg == nil {
		cfg = &config.Config{}
	}
	viper.Unmarshal(cfg)
}

// Execute executes the root command
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// VersionCmd creates the version command
func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Show version information for flat.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("flat version %s\n", version.Version)
			fmt.Printf("Commit: %s\n", version.Commit)
			fmt.Printf("Built: %s\n", version.Date)
		},
	}
}
