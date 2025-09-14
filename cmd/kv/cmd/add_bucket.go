package cmd

import (
	"fmt"
	"os"

	"github.com/yousysadmin/kv/internal/enckeystore"
	"github.com/yousysadmin/kv/internal/storage"

	"github.com/spf13/cobra"
)

var generateNewKey bool

// addKeyCmd represents the add key command
var addBucketCmd = &cobra.Command{
	Use:   "bucket",
	Short: "Add bucket.",
	Long: `This command add a new bucket to the store  and stores.

Arguments:
  <bucket_nam> [flags]

  <bucket_name> A bucket name.`,
	Example: `
  kv add bucket prod-secrets
  kv add bucket prod-secret --generate-new-key # for create a bucket with a separate encryption key`,
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		bucket := args[0]

		s := storage.NewEntityStorage(kvdb, "")

		// Check bucket exist
		if exist, err := s.BucketExist(bucket); err != nil || exist {
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("bucket %q exist\n", bucket)
			os.Exit(1)
		}

		// Create a new bucket
		if err := s.AddBucket(bucket); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// if generate-mew-key is true
		// generate and save new encryption key to the encryption store
		if generateNewKey {
			if encryptionKenStore.HasKey(bucket) {
				fmt.Printf("Encryption key for the bucket %s is exist", bucket)
			}

			enckey, err := enckeystore.GenerateEncryptionKey()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if err := encryptionKenStore.AddKey(bucket, enckeystore.EncryptionKey(enckey)); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if err := encryptionKenStore.Save(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		fmt.Printf("add bucket: %s successfully\n", bucket)
	},
}

func init() {
	addCmd.AddCommand(addBucketCmd)

	addBucketCmd.PersistentFlags().BoolVar(&generateNewKey, "generate-new-key", false, "Generate a new encryption key for this bucket")
}
