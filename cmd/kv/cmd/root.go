package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yousysadmin/kv/internal/storage"
	"go.etcd.io/bbolt"
)

var (
	encryptionKey     string
	db                *bbolt.DB
	defaultBucketName string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kv",
	Short: "Encrypted key-value store",
	Long: `kv is a secure, key-value store for managing encrypted secrets.

You can use --encryption-key or --encryption-key-file to provide an AES key for encryption.
If not provided, a key will be generated automatically and stored in a file.

The database path can be customized with the --db flag or the KV_DB_PATH environment variable.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		k, err := getEncryptKey(viper.GetString("encryption-key"), viper.GetString("encryption-key-file"))
		if err != nil {
			return fmt.Errorf("get encryption key: %w", err)
		}
		encryptionKey = k

		kvdb, err := bbolt.Open(viper.GetString("db"), 0600, nil)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		db = kvdb
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("db", expandPath("~/.kv.db"), "path to database file (can also use KV_DB_PATH)")
	rootCmd.PersistentFlags().String("encryption-key", "", "encryption key (can also use KV_ENCRYPTION_KEY)")
	rootCmd.PersistentFlags().String("encryption-key-file", expandPath("~/.kv.key"), "path to encryption key file (can also use KV_ENCRYPTION_KEY_FILE)")
	rootCmd.PersistentFlags().StringVar(&defaultBucketName, "default-bucket", storage.DefaultBucket, "default bucket name (can also use KV_DEFAULT_BUCKET)")

	viper.BindPFlag("db", rootCmd.PersistentFlags().Lookup("db"))
	viper.BindPFlag("encryption-key", rootCmd.PersistentFlags().Lookup("encryption-key"))
	viper.BindPFlag("encryption-key-file", rootCmd.PersistentFlags().Lookup("encryption-key-file"))
	viper.BindPFlag("default-bucket", rootCmd.PersistentFlags().Lookup("default-bucket"))

	viper.BindEnv("db", "KV_DB_PATH")
	viper.BindEnv("encryption-key", "KV_ENCRYPTION_KEY")
	viper.BindEnv("encryption-key-file", "KV_ENCRYPTION_KEY_FILE")
	viper.BindEnv("default-bucket", "KV_DEFAULT_BUCKET")

	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
		_ = viper.BindPFlags(rootCmd.PersistentFlags())
	})
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}
