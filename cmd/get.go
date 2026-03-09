package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/dmdhrumilmistry/mackey/keystore"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Retrieve a value from the macOS Keychain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		val, err := keystore.Get(service, key)
		if err != nil {
			if errors.Is(err, keystore.ErrNotFound) {
				fmt.Fprintf(os.Stderr, "Key %q not found in service %q\n", key, service)
			} else {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			return err
		}
		fmt.Println(val)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
