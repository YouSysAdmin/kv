package cmd

import (
	"fmt"
	"os"

	"github.com/yousysadmin/kv/internal/storage"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <key>|<key@bucket>",
	Short: "Retrieve a value by key.",
	Long: `This command fetches and decrypts the value associated with the specified key.
If a bucket is not specified, the default bucket will be used.`,
	Example: `
  kv get username
  kv get username@production`,
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		k, b := parseKey(args[0])

		encKey, err := selectKey(encryptionKeys, b)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		s := storage.NewEntityStorage(kvdb, encKey)
		v, err := s.Get(b, k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "get key: `%s` failed: %s\n", k, err.Error())
			os.Exit(1)
		}
		fmt.Printf("%s", v)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
