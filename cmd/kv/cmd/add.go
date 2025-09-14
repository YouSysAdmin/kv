package cmd

import (
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add keys or buckets to the store",
	Long:  "The add command adding keys or buckets to the store.",
	Example: `
  kv add key ...
  kv add bucket prod`,
	Args: cobra.NoArgs,
	//Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
