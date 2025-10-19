package cli

import (
	"github.com/spf13/cobra"
)

var (
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

	importCmd.PersistentFlags().BoolVarP(&importDryRun, "dry-run", "d", false, "Show what would be imported without writing")
	importCmd.PersistentFlags().BoolVarP(&importShowValues, "values", "v", false, "Show key values during import")
}
