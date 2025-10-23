package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yousysadmin/kv/internal/enckeystore"
	"github.com/yousysadmin/kv/internal/storage"
	"go.etcd.io/bbolt"
)

var (
	encryptionKeys     map[string]string
	encryptionKenStore *enckeystore.EncryptionKeyStore
	kvdb               *bbolt.DB
	bucketName         string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kv",
	Short: "Encrypted key-value store",
	Long: `kv is a secure, key-value store for managing encrypted secrets.

You can use --encryption-key or --encryption-key-store-path to provide an AES key for encryption.
If not provided, a key will be generated automatically and stored in a file.

The database path can be customized with the --db flag or the KV_DB_PATH environment variable.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		k, ks, err := loadAllKeys(
			viper.GetString("encryption-key-store"), // path to keys.yaml
			viper.GetString("encryption-key"),       // override encryption key if set via cli flag
		)
		if err != nil {
			return fmt.Errorf("load keys: %w", err)
		}
		encryptionKeys = k
		encryptionKenStore = ks

		kvdb, err = bbolt.Open(viper.GetString("db"), 0o600, nil)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
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
	rootCmd.PersistentFlags().String("encryption-key-store-path", expandPath("~/.kv.key"), "path to encryption key file (can also use KV_ENCRYPTION_KEY_STORE_PATH)")
	rootCmd.PersistentFlags().StringP("bucket", "b", storage.DefaultBucket, "default bucket name (can also use KV_BUCKET)")

	viper.BindPFlag("db", rootCmd.PersistentFlags().Lookup("db"))
	viper.BindPFlag("encryption-key", rootCmd.PersistentFlags().Lookup("encryption-key"))
	viper.BindPFlag("encryption-key-store", rootCmd.PersistentFlags().Lookup("encryption-key-store-path"))
	viper.BindPFlag("bucket", rootCmd.PersistentFlags().Lookup("bucket"))

	viper.BindEnv("db", "KV_DB_PATH")
	viper.BindEnv("encryption-key", "KV_ENCRYPTION_KEY")
	viper.BindEnv("encryption-key-store-path", "KV_ENCRYPTION_KEY_STORE_PATH")
	viper.BindEnv("bucket", "KV_BUCKET")

	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
		_ = viper.BindPFlags(rootCmd.PersistentFlags())

		// bucket name value
		bucketName = viper.GetString("bucket")
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
