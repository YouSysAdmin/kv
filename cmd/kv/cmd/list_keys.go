package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yousysadmin/kv/internal/models"
	"github.com/yousysadmin/kv/internal/storage"

	"github.com/spf13/cobra"
)

var withValues bool
var asJson bool

// listKeysCmd represents the keys command
var listKeysCmd = &cobra.Command{
	Use:   "keys [<bucket>]",
	Short: "List keys.",
	Long: `This command outputs all key names in the current or specified bucket.
It does not display values, only the stored keys.`,
	Example: `
  kv list keys
  kv list keys mybucket`,
	Run: func(cmd *cobra.Command, args []string) {
		var bucket string
		if len(args) == 0 {
			bucket = defaultBucketName
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
	listKeysCmd.PersistentFlags().BoolVar(&withValues, "values", false, "decrypt and output values")
	listKeysCmd.PersistentFlags().BoolVar(&asJson, "json", false, "output as json")
}

// outputKeyList print list of keys in plaintext or json format
func outputKeyList(data []models.Entity) error {
	if asJson {
		if data, err := json.Marshal(data); err == nil {
			fmt.Println(string(data))
		} else {
			return err
		}
	} else {
		for _, kv := range data {
			if withValues {
				fmt.Printf("%s:%s\n", kv.Key, kv.Value)
			} else {
				fmt.Println(kv.Key)
			}
		}
	}
	return nil
}
