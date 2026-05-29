package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	buildVersion string
	buildCommit  string
	buildDate    string
)

var rootCmd = &cobra.Command{
	Use:   "m2p",
	Short: "Convert Markdown files to styled PDFs",
	Long: `m2p converts Markdown files to PDF using a custom typographic style.

Requires Chrome or Chromium installed for PDF rendering.

Environment variables are prefixed with M2P_ (e.g. M2P_PAPER).
Config file: ~/.config/m2p/config.toml (or --config to override).`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute is the entry point called from main.
func Execute(version, commit, date string) {
	buildVersion = version
	buildCommit = commit
	buildDate = date

	if err := rootCmd.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("config", "", "config file path (default: ~/.config/m2p/config.toml)")
	rootCmd.PersistentFlags().BoolP(KeyQuiet, "q", false, "suppress progress output")
	_ = viper.BindPFlag(KeyQuiet, rootCmd.PersistentFlags().Lookup(KeyQuiet))

	rootCmd.CompletionOptions.HiddenDefaultCmd = false

	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
}

func initConfig() {
	cfgFile, _ := rootCmd.Flags().GetString("config")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(ConfigFileName)
		viper.SetConfigType(ConfigFileType)
		viper.AddConfigPath(".")
		viper.AddConfigPath(xdgConfigDir())
	}

	viper.SetEnvPrefix(EnvPrefix)
	viper.AutomaticEnv()

	// Ignore missing config file — it's optional.
	_ = viper.ReadInConfig()
}

func xdgConfigDir() string {
	if runtime.GOOS == OSWindows {
		if d := os.Getenv(EnvAPPDATA); d != "" {
			return filepath.Join(d, ConfigDirName)
		}
	}
	if d := os.Getenv(EnvXDGConfigHome); d != "" {
		return filepath.Join(d, ConfigDirName)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ConfigPathDefault, ConfigDirName)
}

func quietMode() bool {
	return viper.GetBool(KeyQuiet)
}

func infof(format string, args ...any) {
	if !quietMode() {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}
