package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ralvarezdev/m2p/internal/converter"
)

var convertCmd = &cobra.Command{
	Use:   "convert <input.md|dir>",
	Short: "Convert a Markdown file (or folder) to PDF (or HTML)",
	Long: `Convert a Markdown file to a styled PDF document.
Pass a directory to batch-convert all .md files found inside it.

The output file defaults to the input filename with a .pdf extension.
Use --format html to produce a standalone HTML file instead, or
--format both to produce both PDF and HTML at once.

Examples:
  m2p convert README.md
  m2p convert README.md -o out/readme.pdf
  m2p convert README.md --format html
  m2p convert README.md --format both --paper letter
  m2p convert ./docs/`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]

		fi, err := os.Stat(input)
		if err != nil {
			return fmt.Errorf("input not found: %s", input)
		}

		format, err := converter.ParseFormat(viper.GetString(KeyFormat))
		if err != nil {
			return err
		}

		paper, err := converter.ParsePaper(viper.GetString(KeyPaper))
		if err != nil {
			return err
		}

		engine, err := converter.ParseEngine(viper.GetString(KeyEngine))
		if err != nil {
			return err
		}

		noFooter := viper.GetBool(KeyNoFooter)

		pageBreakLevel, err := converter.ParsePageBreak(viper.GetString(KeyPageBreak))
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return runBatch(input, format, paper, engine, noFooter, pageBreakLevel)
		}

		if !strings.EqualFold(filepath.Ext(input), converter.ExtMarkdown) {
			return fmt.Errorf("input must be a %s file, got: %s", converter.ExtMarkdown, filepath.Ext(input))
		}

		openAfter := viper.GetBool(KeyOpen)
		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			output = converter.DefaultOutput(input, format)
		}

		infof("converting %s → %s [%s, %s, engine:%s]", input, output, format, paper, engine)

		opts := converter.Options{
			Input:          input,
			Output:         output,
			Format:         format,
			Paper:          paper,
			Engine:         engine,
			ShowFooter:     !noFooter,
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

func runBatch(
	dir string,
	format converter.Format,
	paper converter.Paper,
	engine converter.Engine,
	noFooter bool,
	pageBreakLevel int,
) error {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.EqualFold(filepath.Ext(path), converter.ExtMarkdown) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk directory: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no .md files found in %s", dir)
	}

	infof("found %d .md file(s) in %s", len(files), dir)
	for _, f := range files {
		output := converter.DefaultOutput(f, format)
		infof("converting %s → %s", f, output)
		opts := converter.Options{
			Input:          f,
			Output:         output,
			Format:         format,
			Paper:          paper,
			Engine:         engine,
			ShowFooter:     !noFooter,
			PageBreakLevel: pageBreakLevel,
		}
		if err := converter.Convert(opts); err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
		infof("done → %s", output)
	}
	return nil
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
