package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yousysadmin/kv/internal/storage"

	"github.com/spf13/cobra"
)

// addKeyCmd represents the add key command
var addKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Add or update a key-value pair.",
	Long: `This command encrypts and stores the provided value using the configured AES key.
You can specify the encryption key via the --encryption-key flag.

Arguments:
  <key>|<key@bucket> <value>

  <key>         The key to store the value under in the default bucket.
  <key@bucket>  The key to store in a specific named bucket.
  <value>       The value to be encrypted and stored.`,
	Example: `
  kv add key username admin
  kv add key --bucket=prod username admin
  kv add key password@auth supersecret
  kv add key config@prod '{"debug":false}'
  # read value from file
  kv add key longtext @readme.txt
  # read value from STDIN
  echo 'env=prod' | kv add key config@env @-`,
	Args: cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		k, b := parseKey(args[0])

		encKey, err := selectKey(encryptionKeys, b)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		s := storage.NewEntityStorage(kvdb, encKey)
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
	addCmd.AddCommand(addKeyCmd)
}
