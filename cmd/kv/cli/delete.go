package cli

import (
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete keys or buckets from the store",
	Long: `The delete command removes keys or entire buckets from the store.

Keys can be specified directly or with a bucket using the key@bucket syntax.
Deleting a bucket will remove all keys under that bucket.

This action is permanent and cannot be undone.`,
	Example: `
  kv delete key mykey
  kv delete key mykey@prod
  kv delete bucket prod`,
	Args: cobra.NoArgs,
	//Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
