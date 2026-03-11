package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dmdhrumilmistry/mackey/keystore"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec -- <command> [args...]",
	Short: "Run a command with keychain secrets injected as environment variables",
	Long: `exec runs the given command as a child process with all stored key-value
pairs for the service injected into its environment. Secrets are never written
to disk and are not visible in the parent shell's environment.

All subprocesses spawned by the child command inherit the same environment, so
secrets are available throughout the entire process tree.

Shell expansion note:
  Variables like $SECRET_KEY are expanded by your shell BEFORE mackey runs, so
  the shell substitutes an empty string when the variable is not set in the
  parent shell. To reference injected secrets inside the command, wrap it in a
  shell invocation using single quotes:

    mackey exec -- sh -c 'echo $SECRET_KEY'

Examples:

  # Run an application that reads env vars at startup
  mackey exec -- node server.js
  mackey exec -- python manage.py runserver
  mackey exec --service myapp -- docker-compose up

  # Print a specific secret (no shell expansion issues)
  mackey exec -- printenv SECRET_KEY

  # Use shell features that reference injected variables
  mackey exec -- sh -c 'echo $SECRET_KEY'
  mackey exec -- sh -c 'curl -H "Authorization: Bearer $API_TOKEN" https://api.example.com'`,
	// At least the command to run must be provided.
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		env, err := buildExecEnv(service, os.Environ())
		if err != nil {
			return err
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

// buildExecEnv constructs the environment slice for the child process.
// It starts from parentEnv (typically os.Environ()), strips any variable that
// is also present in the keychain for svc, and then appends the keychain
// values so they always take effect.  This ensures keychain secrets override
// any identically-named variable that may exist in the caller's shell.
func buildExecEnv(svc string, parentEnv []string) ([]string, error) {
	keys, err := keystore.List(svc)
	if err != nil {
		return nil, fmt.Errorf("error listing keys: %w", err)
	}

	// Fetch all keychain values first so we know which names to override.
	keychain := make(map[string]string, len(keys))
	for _, key := range keys {
		val, err := keystore.Get(svc, key)
		if err != nil {
			if errors.Is(err, keystore.ErrNotFound) {
				continue
			}
			return nil, fmt.Errorf("error reading key %s: %w", key, err)
		}
		keychain[key] = val
	}

	// Inherit the parent environment (PATH, HOME, etc.) but drop any entry
	// whose name is overridden by a keychain value so there are no
	// duplicates. On POSIX systems getenv(3) returns the *first* match, so
	// a trailing duplicate would never be seen by the child process.
	env := make([]string, 0, len(parentEnv)+len(keychain))
	for _, e := range parentEnv {
		name, _, _ := strings.Cut(e, "=")
		if _, override := keychain[name]; !override {
			env = append(env, e)
		}
	}
	for k, v := range keychain {
		env = append(env, k+"="+v)
	}
	return env, nil
}

func init() {
	rootCmd.AddCommand(execCmd)
}
