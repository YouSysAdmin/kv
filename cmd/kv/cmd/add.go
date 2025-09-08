package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yousysadmin/kv/internal/storage"

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
  kv add config@prod '{"debug":false}'
  kv add longtext @readme.txt
  echo 'env=prod' | kv add config@env @-`,
	Args: cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		k, b := parseKey(args[0])
		s := storage.NewEntityStorage(db, encryptionKey)
		val, err := readValue(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "add key: %s failed: %s\n", k, err.Error())
			os.Exit(1)
		}
		err = s.Add(b, k, val)
		if err != nil {
			fmt.Fprintf(os.Stderr, "add key: %s failed: %s\n", k, err.Error())
			os.Exit(1)
		}

		fmt.Printf("add key: %s successfully\n", k)
	},
}

// readValue interprets a value argument, supporting:
// plain string
// @filename to read content from a file
// @- to read content from stdin
func readValue(arg string) (string, error) {
	if strings.HasPrefix(arg, "@") {
		if arg == "@-" {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return "", fmt.Errorf("failed to read from stdin: %w", err)
			}
			return string(data), nil
		}
		data, err := os.ReadFile(arg[1:])
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", arg[1:], err)
		}
		return string(data), nil
	}
	return arg, nil
}

func init() {
	rootCmd.AddCommand(addCmd)
}
