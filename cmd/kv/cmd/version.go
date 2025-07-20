package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/yousysadmin/kv/pkg"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("kv: %s\n", pkg.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
