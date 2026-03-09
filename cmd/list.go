package cmd

import (
	"fmt"
	"os"

	"github.com/dmdhrumilmistry/mackey/keystore"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all keys stored in the macOS Keychain for the given service",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		keys, err := keystore.List(service)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return err
		}
		if len(keys) == 0 {
			fmt.Fprintf(os.Stderr, "No keys found for service %q\n", service)
			return nil
		}
		for _, k := range keys {
			fmt.Println(k)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
