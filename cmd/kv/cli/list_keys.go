package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/yousysadmin/kv/internal/models"
	"github.com/yousysadmin/kv/internal/storage"
	"github.com/yousysadmin/kv/internal/utils"

	"github.com/spf13/cobra"
)

var (
	withValues bool
	format     string
)

// listKeysCmd represents the keys command
var listKeysCmd = &cobra.Command{
	Use:   "keys [<bucket>]",
	Short: "List keys.",
	Long: `This command outputs all key names in the current or specified bucket.
It does not display values, only the stored keys.`,
	Example: `
  kv list keys
  kv list keys mybucket
  kv list keys --bucket=mybucket`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		s := storage.NewEntityStorage(kvdb, "")
		bl, err := s.ListBuckets()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		var out []string
		for _, k := range bl {
			if strings.HasPrefix(k, toComplete) {
				out = append(out, k)
			}
		}
		return out, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		var bucket string
		if len(args) == 0 {
			bucket = bucketName
		} else {
			bucket = args[0]
		}

		encKey, err := selectKey(encryptionKeys, bucket)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		s := storage.NewEntityStorage(kvdb, encKey)
		v, err := s.List(bucket, withValues)
		if err != nil {
			fmt.Fprintf(os.Stderr, "list keys in bucket: `%s` failed: %s\n", bucket, err.Error())
			os.Exit(1)
		}
		if err := outputKeyList(v); err != nil {
			fmt.Fprintf(os.Stderr, "list keys in bucket: `%s` failed: %s\n", bucket, err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	listCmd.AddCommand(listKeysCmd)
	listKeysCmd.PersistentFlags().BoolVarP(&withValues, "values", "v", false, "decrypt and output values")
	listKeysCmd.PersistentFlags().StringVarP(&format, "format", "f", "raw", "output format [raw, json, dotenv, rails-dotenv]")
}

// outputKeyList print list of keys in plaintext or json format
func outputKeyList(data []models.Entity) error {
	switch format {
	case "raw":
		if err := printRaw(data, withValues); err != nil {
			return err
		}
	case "dotenv", "rails-dotenv":
		if err := printDotenv(data, format, withValues); err != nil {
			return err
		}
	case "json":
		if err := printJson(data); err != nil {
			return err
		}
	default:
		return errors.New("unkown output format")

	}

	return nil
}

// printDotenv print list of key as dotenv
func printDotenv(data []models.Entity, format string, withValues bool) error {
	var dotenvFormat utils.DotenvMode

	if format == "rails-dotenv" {
		dotenvFormat = utils.DotenvMultiline
	} else {
		dotenvFormat = utils.DotenvEscaped
	}
	o, err := utils.ToDotenvMode(data, withValues, dotenvFormat)
	if err != nil {
		return err
	}
	fmt.Println(o)
	return nil

}

// printJson prinv list of kv as json
func printJson(data []models.Entity) error {
	if data, err := json.Marshal(data); err == nil {
		fmt.Println(string(data))
	} else {
		return err
	}
	return nil
}

// printRaw print kv as raw
func printRaw(data []models.Entity, withValues bool) error {
	for _, kv := range data {
		if withValues {
			fmt.Printf("%s:%s\n", kv.Key, kv.Value)
		} else {
			fmt.Println(kv.Key)
		}
	}

	return nil
}
