package cmd

import (
	"fmt"
	"github.com/yousysadmin/kv/internal/storage"
	"os"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add <key>|<key@bucket> <value>",
	Short: "Add or update a key-value pair.",
	Long: `This command encrypts and stores the provided value using the configured AES key.
You can specify the encryption key via the --encryption-key flag or the --encryption-key-file option.

Arguments:
  <key>         The key to store the value under in the default bucket.
  <key@bucket>  The key to store in a specific named bucket.
  <value>       The value to be encrypted and stored.`,
	Example: `
  kv add username admin
  kv add password@auth supersecret
  kv add config@prod '{"debug":false}'`,
	Args: cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		k, b := parseKey(args[0])
		s := storage.NewEntityStorage(db, encryptionKey)
		err := s.Add(b, k, args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "add key: %s failed: %s\n", k, err.Error())
			os.Exit(1)
		}

		fmt.Printf("add key: %s successfully\n", k)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
