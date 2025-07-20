package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.etcd.io/bbolt"
	"os"
	"path/filepath"
	"strings"
)

var encryptionKey string
var db *bbolt.DB

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

	viper.BindPFlag("db", rootCmd.PersistentFlags().Lookup("db"))
	viper.BindPFlag("encryption-key", rootCmd.PersistentFlags().Lookup("encryption-key"))
	viper.BindPFlag("encryption-key-file", rootCmd.PersistentFlags().Lookup("encryption-key-file"))

	viper.BindEnv("db", "KV_DB_PATH")
	viper.BindEnv("encryption-key", "KV_ENCRYPTION_KEY")
	viper.BindEnv("encryption-key-file", "KV_ENCRYPTION_KEY_FILE")

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
