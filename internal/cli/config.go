package cli

import (
	"encoding/json"
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
		keys := []string{KeyPaper, KeyFormat, KeyEngine, KeyPageBreak, KeyNoFooter, KeyOpen, KeyQuiet}

		cfgFile := viper.ConfigFileUsed()
		if cfgFile == "" {
			cfgFile = "(none)"
		}

		outputFmt, _ := cmd.Flags().GetString("output")

		if outputFmt == "json" {
			m := make(map[string]any, len(keys)+1)
			m["config_file"] = cfgFile
			for _, k := range keys {
				m[k] = viper.Get(k)
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(m)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "config file: %s\n\n", cfgFile)
		for _, k := range keys {
			fmt.Fprintf(cmd.OutOrStdout(), "  %-12s = %v\n", k, viper.Get(k))
		}
		return nil
	},
}

func init() {
	configShowCmd.Flags().StringP("output", "o", "text", "output format: text, json")
	configCmd.AddCommand(configShowCmd)
}
