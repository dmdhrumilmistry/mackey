// Package cmd implements the mackey CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var service string

var rootCmd = &cobra.Command{
	Use:   "mackey",
	Short: "mackey – secure key-value store backed by the macOS Keychain",
	Long: `mackey stores and retrieves key-value pairs using the macOS Keychain so
that you never need to keep sensitive values in .env files or environment
variables again.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&service, "service", "s", "mackey",
		"Keychain service name used to namespace all keys")
}
