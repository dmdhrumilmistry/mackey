package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/dmdhrumilmistry/mackey/keystore"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete <key>",
	Aliases: []string{"del", "rm"},
	Short:   "Delete a key from the macOS Keychain",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		err := keystore.Delete(service, key)
		if err != nil {
			if errors.Is(err, keystore.ErrNotFound) {
				fmt.Fprintf(os.Stderr, "Key %q not found in service %q\n", key, service)
			} else {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			return err
		}
		fmt.Printf("Deleted %q from service %q\n", key, service)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
