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

		if !strings.EqualFold(filepath.Ext(input), converter.ExtMarkdown) {
			return fmt.Errorf("input must be a %s file, got: %s", converter.ExtMarkdown, filepath.Ext(input))
		}

		if _, err := os.Stat(input); err != nil {
			return fmt.Errorf("input file not found: %s", input)
		}

		formatStr := viper.GetString(KeyFormat)
		if v, _ := cmd.Flags().GetString(KeyFormat); v != "" {
			formatStr = v
		}
		format, err := converter.ParseFormat(formatStr)
		if err != nil {
			return err
		}

		paperStr := viper.GetString(KeyPaper)
		if v, _ := cmd.Flags().GetString(KeyPaper); v != "" {
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

		engineStr := viper.GetString(KeyEngine)
		if v, _ := cmd.Flags().GetString(KeyEngine); v != "" {
			engineStr = v
		}
		engine, err := converter.ParseEngine(engineStr)
		if err != nil {
			return err
		}

		noFooter, _ := cmd.Flags().GetBool(KeyNoFooter)
		openAfter, _ := cmd.Flags().GetBool(KeyOpen)

		pageBreakStr := viper.GetString(KeyPageBreak)
		if v, _ := cmd.Flags().GetString(KeyPageBreak); v != "" {
			pageBreakStr = v
		}
		pageBreakLevel, err := converter.ParsePageBreak(pageBreakStr)
		if err != nil {
			return err
		}

		infof("converting %s → %s [%s, %s, engine:%s]", input, output, format, paper, engine)

		opts := converter.Options{
			Input:          input,
			Output:         output,
			Format:         format,
			Paper:          paper,
			Engine:         engine,
			ShowFooter:     !noFooter,
			Open:           openAfter,
			PageBreakLevel: pageBreakLevel,
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
	convertCmd.Flags().StringP(KeyFormat, "f", DefaultFormat, "output format: pdf, html, both")
	convertCmd.Flags().String(KeyPaper, DefaultPaper, "paper size: a4, letter, a3, legal")
	convertCmd.Flags().String(KeyEngine, DefaultEngine, "PDF engine: auto, chromium, native")
	convertCmd.Flags().Bool(KeyNoFooter, false, "suppress the brand footer")
	convertCmd.Flags().Bool(KeyOpen, false, "open output file after conversion")
	convertCmd.Flags().String(KeyPageBreak, DefaultPageBreak, "insert page breaks before headings: none, h2, h3")

	// Bind to viper so env vars and config file apply.
	_ = viper.BindPFlag(KeyFormat, convertCmd.Flags().Lookup(KeyFormat))
	_ = viper.BindPFlag(KeyPaper, convertCmd.Flags().Lookup(KeyPaper))
	_ = viper.BindPFlag(KeyEngine, convertCmd.Flags().Lookup(KeyEngine))
	_ = viper.BindPFlag(KeyNoFooter, convertCmd.Flags().Lookup(KeyNoFooter))
	_ = viper.BindPFlag(KeyOpen, convertCmd.Flags().Lookup(KeyOpen))
	_ = viper.BindPFlag(KeyPageBreak, convertCmd.Flags().Lookup(KeyPageBreak))

	viper.SetDefault(KeyFormat, DefaultFormat)
	viper.SetDefault(KeyPaper, DefaultPaper)
	viper.SetDefault(KeyEngine, DefaultEngine)
	viper.SetDefault(KeyPageBreak, DefaultPageBreak)
}
