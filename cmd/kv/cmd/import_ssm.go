package cmd

import (
	"fmt"
	awsSsm "github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/spf13/cobra"
	"github.com/yousysadmin/kv/internal/importer/amazon"
	"github.com/yousysadmin/kv/internal/importer/amazon/ssm"
	"github.com/yousysadmin/kv/internal/storage"
	"os"
)

var (
	ssmAwsAccessKeyID     string
	ssmAwsAccessKeySecret string
	ssmAwsProfileName     string
	ssmAwsRegion          string
	ssmRecursive          bool
	ssmTrimKeyName        bool
)

// importSsmCmd represents the ssm command
var importSsmCmd = &cobra.Command{
	Use:   "ssm",
	Short: "Import secrets from AWS SSM Parameter Store",
	Long:  `Fetch parameters from AWS Systems Manager (SSM) Parameter Store and store them in the local kv bucket. Supports profile or access key authentication, recursive path scanning, and key name trimming.`,
	Example: `
  # Import all secrets under a path using the default profile
  # Region must be set in ~/.aws/config, ~/.aws/credentials or AWS_REGION environment variable
  kv import ssm --bucket=mybucket /prod/secrets

  # Use a specific AWS profile and region
  kv import ssm --bucket=mybucket --profile=dev --region=ca-central-1 /dev/app/config

  # Use static access key credentials (profile is ignored)
  # Can use a default AWS CLI environment variable
  # https://docs.aws.amazon.com/cli/v1/userguide/cli-configure-envvars.html#envvars-set
  kv import ssm --bucket=mybucket --key-id=AKIA... --secret-key=... --region=us-west-2 /prod/secure

  # Recursively import full key names under a prefix
  # e.g., /prod/env/database_password
  kv import ssm --bucket=mybucket --recursive /prod/env/

  # Recursively import and trim key names
  # e.g., /prod/env/database_password => database_password
  kv import ssm --bucket=mybucket --recursive --trim-key-name /prod/env/

  # Perform a dry-run import without writing data
  kv import ssm --bucket=mybucket --dry-run /dev

  # Dry-run and show key values for verification
  kv import ssm --bucket=mybucket --dry-run --show-values /prod
`,
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		if path == "" {
			path = "/"
		}

		if importBucketName == "" {
			fmt.Fprintln(os.Stderr, "import from ssm failed: bucket name is required")
			os.Exit(1)
		}

		ctx := cmd.Context()
		cfg, err := amazon.New(ctx, ssmAwsProfileName, ssmAwsRegion, ssmAwsAccessKeyID, ssmAwsAccessKeySecret)
		if err != nil {
			fmt.Fprintf(os.Stderr, "import from ssm failed: %s\n", err.Error())
			os.Exit(1)
		}
		ssmClient := awsSsm.NewFromConfig(cfg)
		secrets, err := ssm.GetSecrets(ctx, ssmClient, path, ssmRecursive, ssmTrimKeyName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "import from ssm failed: %s\n", err.Error())
			os.Exit(1)
		}

		s := storage.NewEntityStorage(db, encryptionKey)
		for key, value := range secrets {
			if importDryRun {
				if importShowValues {
					fmt.Printf("DryRun[%s]: %s => %s\n", importBucketName, key, value)
				} else {
					fmt.Printf("DryRun[%s]: %s\n", importBucketName, key)
				}

			} else {
				err := s.Add(importBucketName, key, value)
				if err != nil {
					fmt.Fprintf(os.Stderr, "add key: %s failed: %s\n", key, err.Error())
					os.Exit(1)
				}
				if importShowValues {
					fmt.Printf("Imported[%s]: %s => %s\n", importBucketName, key, value)
				} else {
					fmt.Printf("Imported[%s]: %s\n", importBucketName, key)
				}
			}
		}
		fmt.Printf("Imported %d keys.\n", len(secrets))
	},
}

func init() {
	importCmd.AddCommand(importSsmCmd)

	importSsmCmd.PersistentFlags().StringVar(&ssmAwsAccessKeyID, "key-id", "", "AWS Access Key ID")
	importSsmCmd.PersistentFlags().StringVar(&ssmAwsAccessKeySecret, "secret-key", "", "AWS Secret Access Key")
	importSsmCmd.PersistentFlags().StringVar(&ssmAwsProfileName, "profile", "default", "AWS Profile name")
	importSsmCmd.PersistentFlags().StringVar(&ssmAwsRegion, "region", "", "AWS Region")
	importSsmCmd.PersistentFlags().BoolVarP(&ssmTrimKeyName, "trim-key-name", "t", false, "Trim key name to only the final path part (e.g., /foo/bar → bar)")
	importSsmCmd.PersistentFlags().BoolVarP(&ssmRecursive, "recursive", "r", false, "Recursive import")

}
