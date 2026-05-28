package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ralvarez/m2p/internal/converter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var convertCmd = &cobra.Command{
	Use:   "convert <input.md>",
	Short: "Convert a Markdown file to PDF (or HTML)",
	Long: `Convert a Markdown file to a styled PDF document.

The output file defaults to the input filename with a .pdf extension.
Use --format html to produce a standalone HTML file instead, or
--format both to produce both PDF and HTML at once.

Examples:
  m2p convert README.md
  m2p convert README.md -o out/readme.pdf
  m2p convert README.md --format html
  m2p convert README.md --format both --paper letter`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]

		if !strings.EqualFold(filepath.Ext(input), ".md") {
			return fmt.Errorf("input must be a .md file, got: %s", filepath.Ext(input))
		}

		if _, err := os.Stat(input); err != nil {
			return fmt.Errorf("input file not found: %s", input)
		}

		formatStr := viper.GetString("format")
		if v, _ := cmd.Flags().GetString("format"); v != "" {
			formatStr = v
		}
		format, err := converter.ParseFormat(formatStr)
		if err != nil {
			return err
		}

		paperStr := viper.GetString("paper")
		if v, _ := cmd.Flags().GetString("paper"); v != "" {
			paperStr = v
		}
		paper, err := converter.ParsePaper(paperStr)
		if err != nil {
			return err
		}

		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			output = converter.DefaultOutput(input, format)
		}

		engineStr := viper.GetString("engine")
		if v, _ := cmd.Flags().GetString("engine"); v != "" {
			engineStr = v
		}
		engine, err := converter.ParseEngine(engineStr)
		if err != nil {
			return err
		}

		noFooter, _ := cmd.Flags().GetBool("no-footer")
		openAfter, _ := cmd.Flags().GetBool("open")

		infof("converting %s → %s [%s, %s, engine:%s]", input, output, format, paper, engine)

		opts := converter.Options{
			Input:      input,
			Output:     output,
			Format:     format,
			Paper:      paper,
			Engine:     engine,
			ShowFooter: !noFooter,
			Open:       openAfter,
		}

		if err := converter.Convert(opts); err != nil {
			return err
		}

		infof("done → %s", output)

		if openAfter {
			return openFile(output)
		}
		return nil
	},
}

func init() {
	convertCmd.Flags().StringP("output", "o", "", "output file path (default: <input>.pdf)")
	convertCmd.Flags().StringP("format", "f", "pdf", "output format: pdf, html, both")
	convertCmd.Flags().String("paper", "a4", "paper size: a4, letter, a3, legal")
	convertCmd.Flags().String("engine", "auto", "PDF engine: auto, chromium, native")
	convertCmd.Flags().Bool("no-footer", false, "suppress the brand footer")
	convertCmd.Flags().Bool("open", false, "open output file after conversion")

	// Bind to viper so env vars and config file apply.
	_ = viper.BindPFlag("format", convertCmd.Flags().Lookup("format"))
	_ = viper.BindPFlag("paper", convertCmd.Flags().Lookup("paper"))
	_ = viper.BindPFlag("engine", convertCmd.Flags().Lookup("engine"))
	_ = viper.BindPFlag("no-footer", convertCmd.Flags().Lookup("no-footer"))
	_ = viper.BindPFlag("open", convertCmd.Flags().Lookup("open"))

	viper.SetDefault("format", "pdf")
	viper.SetDefault("paper", "a4")
	viper.SetDefault("engine", "auto")
}
