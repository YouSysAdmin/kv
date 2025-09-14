package cmd

import (
	"fmt"
	"os"

	"github.com/yousysadmin/kv/internal/storage"

	"github.com/spf13/cobra"
)

// listBucketsCmd represents the listBuckets command
var listBucketsCmd = &cobra.Command{
	Use:   "buckets",
	Short: "List buckets",
	Long: `List all available buckets in the encrypted key-value database.

This command retrieves and prints the names of all top-level buckets stored in the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		s := storage.NewEntityStorage(kvdb, "")
		bl, err := s.ListBuckets()
		if err != nil {
			fmt.Fprintf(os.Stderr, "list bucketets: failed: %s\n", err.Error())
			os.Exit(1)
		}
		for _, b := range bl {
			fmt.Println(b)
		}
	},
}

func init() {
	listCmd.AddCommand(listBucketsCmd)
}
