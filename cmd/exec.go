package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/dmdhrumilmistry/mackey/keystore"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec -- <command> [args...]",
	Short: "Run a command with keychain secrets injected as environment variables",
	Long: `exec runs the given command as a child process with all stored key-value
pairs for the service injected into its environment. Secrets are never written
to disk and are not visible in the parent shell's environment.

Examples:

  mackey exec -- node server.js
  mackey exec -- python app.py
  mackey exec --service myapp -- docker-compose up`,
	// At least the command to run must be provided.
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keys, err := keystore.List(service)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error listing keys:", err)
			return err
		}

		// Start with the current process environment so the child inherits
		// existing variables (PATH, HOME, etc.) and keychain values are layered on top.
		env := os.Environ()
		for _, key := range keys {
			val, err := keystore.Get(service, key)
			if err != nil {
				if errors.Is(err, keystore.ErrNotFound) {
					continue
				}
				fmt.Fprintln(os.Stderr, "Error reading key:", key, err)
				return err
			}
			env = append(env, key+"="+val)
		}

		child := exec.Command(args[0], args[1:]...) //nolint:gosec // intentional: user controls the command
		child.Env = env
		child.Stdin = os.Stdin
		child.Stdout = os.Stdout
		child.Stderr = os.Stderr

		if err := child.Run(); err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				// Propagate the child's exit code rather than printing an error.
				os.Exit(exitErr.ExitCode())
			}
			fmt.Fprintln(os.Stderr, "Error:", err)
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}
