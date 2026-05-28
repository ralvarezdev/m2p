package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage m2p configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show resolved configuration and value sources",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		keys := []string{KeyPaper, KeyFormat, KeyNoFooter, KeyOpen, KeyQuiet}

		cfgFile := viper.ConfigFileUsed()
		if cfgFile == "" {
			cfgFile = "(none)"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "config file: %s\n\n", cfgFile)

		for _, k := range keys {
			fmt.Fprintf(cmd.OutOrStdout(), "  %-12s = %v\n", k, viper.Get(k))
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
}
