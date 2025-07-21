package cmd

import (
	"github.com/spf13/cobra"
)

var (
	importBucketName string
	importDryRun     bool
	importShowValues bool
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import secrets from external store",
	Long:  ``,
	Args:  cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.PersistentFlags().StringVarP(&importBucketName, "bucket", "b", "", "Target bucket name")
	importCmd.PersistentFlags().BoolVar(&importDryRun, "dry-run", false, "Show what would be imported without writing")
	importCmd.PersistentFlags().BoolVar(&importShowValues, "show-values", false, "Show key values during import")
}
