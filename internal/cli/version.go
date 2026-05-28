package cli

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

type versionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
	Go      string `json:"go"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		outputFmt, _ := cmd.Flags().GetString("output")

		info := versionInfo{
			Version: buildVersion,
			Commit:  buildCommit,
			Date:    buildDate,
			Go:      runtime.Version(),
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
		}

		if outputFmt == "json" {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(info)
		}

		fmt.Fprintf(cmd.OutOrStdout(),
			"m2p %s (rev %s, built %s, %s)\n",
			info.Version, info.Commit, info.Date, info.Go,
		)
		return nil
	},
}

func init() {
	versionCmd.Flags().StringP("output", "o", "text", "output format: text, json")
}
