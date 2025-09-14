package cmd

import (
	"fmt"
	"os"

	"github.com/yousysadmin/kv/internal/storage"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Delete a key.",
	Long: `Delete a key from the store.

This command removes the specified key from store.
If no bucket is provided, the key will be deleted from the default bucket.`,
	Example: `
  kv delete username
  kv delete token@authservice`,
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		k, b := parseKey(args[0])
		s := storage.NewEntityStorage(kvdb, "")
		err := s.Delete(b, k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "delete key: %s in bucket %s failed: %s\n", k, b, err.Error())
			os.Exit(1)
		}

		fmt.Printf("delete key: %s in bucket %s successfully\n", k, b)
	},
}

func init() {
	deleteCmd.AddCommand(deleteKeyCmd)
}
