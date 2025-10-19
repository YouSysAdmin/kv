package cli

import (
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List keys/buckets.",
	Long:  ``,
	Example: `
  kv list keys
  kv list keys mybucket
  kv list buckets`,
	Args: cobra.NoArgs,
	//Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
