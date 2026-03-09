package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/dmdhrumilmistry/mackey/keystore"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Print stored key-value pairs as shell export statements",
	Long: `env prints all key-value pairs for the given service as shell export
statements so that you can source them into your current shell:

  eval "$(mackey env)"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		keys, err := keystore.List(service)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return err
		}

		for _, key := range keys {
			val, err := keystore.Get(service, key)
			if err != nil {
				if errors.Is(err, keystore.ErrNotFound) {
					continue
				}
				fmt.Fprintln(os.Stderr, "Error reading key:", key, err)
				return err
			}
			// Single-quote the value to prevent shell expansion of its contents.
			fmt.Printf("export %s='%s'\n", key, shellEscape(val))
		}
		return nil
	},
}

// shellEscape escapes single-quote characters in a value for safe shell quoting.
func shellEscape(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\'' {
			// End the single-quoted string, add an escaped single quote, restart.
			out = append(out, '\'', '\\', '\'', '\'')
		} else {
			out = append(out, s[i])
		}
	}
	return string(out)
}

func init() {
	rootCmd.AddCommand(envCmd)
}
