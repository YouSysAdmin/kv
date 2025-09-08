package cmd

import (
	"fmt"
	"os"

	"github.com/yousysadmin/kv/internal/storage"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteBucketCmd = &cobra.Command{
	Use:   "bucket",
	Short: "Delete a bucket.",
	Long: `Delete a bucket from the store.

This command removes the specified bucket from the store.`,
	Example: `
  kv delete prod.`,
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		bucket := args[0]
		s := storage.NewEntityStorage(db, encryptionKey)
		err := s.DeleteBucket(bucket)
		if err != nil {
			fmt.Fprintf(os.Stderr, "delete bucket: %s failed: %s\n", bucket, err.Error())
			os.Exit(1)
		}

		fmt.Printf("delete bucket: %s successfully\n", bucket)
	},
}

func init() {
	deleteCmd.AddCommand(deleteBucketCmd)
}
