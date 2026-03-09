package cmd

import (
	"fmt"
	"os"

	"github.com/dmdhrumilmistry/mackey/keystore"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Store a key-value pair in the macOS Keychain",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		if err := keystore.Set(service, key, value); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return err
		}
		fmt.Printf("Stored %q under service %q\n", key, service)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}
